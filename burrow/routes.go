package burrow

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Pattern string
	Handler HTTPHandler
}

type RouteHandler struct {
	Prefix string
	Routes []Route
}

func (h *RouteHandler) Subrouter(r *mux.Router) (subrouter *mux.Router) {
	subrouter = r.PathPrefix(h.Prefix).Subrouter()
	for _, route := range h.Routes {
		subrouter.Handle(route.Pattern, route.Handler)
	}
	return
}

func (h *RouteHandler) MountMethods(methods []Method) {
	for _, m := range methods {
		if m.Route != "" {
			method := m
			h.Routes = append(h.Routes, Route{
				Pattern: method.Route,
				Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
					req := &RequestMessage{
						Params: VarsInterface(mux.Vars(r)), //TODO get url params and merge w/ ivars
						Method: &method.Name,
					}
					ctx := &RequestContext{
						HTTPWriter: w,
						HTTPReader: r,
						Request:    req,
					}
					return method.Execute(ctx)
				},
			})
		}
	}
}

func (h *RouteHandler) MountHelp(methods map[string]Method) {
	//TODO include route help w/ order
	helpRoute := Route{
		Pattern: "/help",
		Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
			helper := make(map[string][]string)
			for _, method := range methods {
				helper["methods"] = append(helper["methods"], method.Name)
				if method.Route == "" {
					helper["io_methods"] = append(helper["io_methods"], method.Name)
				}
			}
			helper["info"] = append(helper["info"], "Use help/{method} for method help")
			helper["info"] = append(helper["info"], "io_methods are only usable through websockets")
			return SuccessMsg(helper)
		},
	}
	subhelpRoute := Route{
		Pattern: "/help/{method}",
		Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
			name := mux.Vars(r)["method"]
			if method, ok := methods[name]; !ok {
				return cerrorf(RpcMethodNotFound, "The limit does not exist! %s", name).ResponseMessage()
			} else {
				return SuccessMsg(method)
			}
		},
	}
	h.Routes = append(h.Routes, []Route{subhelpRoute, helpRoute}...)
}

type HTTPHandler func(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage)

func (handle HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// info("Request - %v", r)
	if r.TLS == nil {
		info("NO TLS")
	}
	response := handle(w, r)
	if response != nil {
		content, err := response.Marshal()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if response.Error != nil {
			http.Error(w, string(content), response.Error.Code)
		} else {
			w.Header().Set("Content-Length", sprintSizeOf(content))
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			w.Write(content)
		}
	}
}

// From http://www.jsonrpc.org/specification
// Content-Type: MUST be application/json.
// Content-Length: MUST contain the correct length according to the HTTP-specification.
// Accept: MUST be application/json.

func (h *RouteHandler) MountRpc(ms map[string]Method) {
	methods := ms
	route := Route{
		Pattern: "/rpc",
		Handler: func(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
			switch {
			case r.Method != "POST":
				msg = cerrorf(http.StatusMethodNotAllowed, "Requires method: POST").ResponseMessage()
			case r.Header.Get("Content-Type") != "application/json":
				msg = cerrorf(http.StatusUnsupportedMediaType, "Requires Content-Type: application/json").ResponseMessage()
			case r.Header.Get("Accept") != "application/json":
				msg = cerrorf(http.StatusNotAcceptable, "Requires Accept: application/json").ResponseMessage()
			case r.Header.Get("Content-Length") == "":
				//TODO is it necessary to asset lenght is correct?
				msg = cerrorf(http.StatusLengthRequired, "Requires valid Content-Length").ResponseMessage()
			default:
				if req, rerr := ReadRequestMessage(r.Body); rerr != nil {
					msg = cerrorf(rerr.Code, rerr.Message).ResponseMessage()
				} else {
					ctx := &RequestContext{
						HTTPWriter: w,
						HTTPReader: r,
						Request:    req,
					}
					if method, ok := methods[ctx.Request.MethodName()]; !ok {
						msg = cerrorf(RpcMethodNotFound, "The method does not exist! %s", method).ResponseMessage()
					} else {
						msg = method.Execute(ctx)
					}
				}
			}
			return
		},
	}
	h.Routes = append(h.Routes, route)
}

func (h *RouteHandler) MountIo(ms map[string]Method) {
	methods := ms
	route := Route{
		Pattern: "/io",
		Handler: func(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
			if r.Method != "GET" {
				msg = cerrorf(http.StatusMethodNotAllowed, "Only GET can be upgraded").ResponseMessage()
				return
			}
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				msg = cerrorf(http.StatusBadRequest, "Request can't be upgraded").ResponseMessage()
				return
			}
			c := NewConnection(ws)
			if cerr := c.listen(methods); cerr != nil {
				msg = cerr.ResponseMessage()
			} else {
				msg = SuccessMsg("WS connection closed succesfully")
			}
			//--> SUBSCRIBE on tile load -> {tileset, tileXYZ}
			//--> UNSUBSCRIBE on tile unload -> {tileset, tileXYZ}
			//--> LIST_SUBSCRIPTIONS
			//<-- {tileset, tile, data, type}
			msg = nil // Can't return a body...
			return
		},
	}
	h.Routes = append(h.Routes, route)
}

type RequestContext struct {
	Request    *RequestMessage
	Connection *Connection
	HTTPWriter http.ResponseWriter
	HTTPReader *http.Request
	Params     MethodParams
}

func (r *RouteHandler) MountRoutes(methods []Method) {
	dict := make(map[string]Method)
	for _, method := range methods {
		dict[method.Name] = method
	}
	r.MountIo(dict)
	r.MountRpc(dict)
	r.MountMethods(methods) //preserve order
	r.MountHelp(dict)
}

func VarsInterface(vars map[string]string) map[string]interface{} {
	ivars := make(map[string]interface{})
	for k, v := range vars {
		// cast nums to float64s
		if fv, err := atof(v); err == nil {
			ivars[k] = fv
		} else {
			ivars[k] = v
		}
	}
	return ivars
}
