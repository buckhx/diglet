package digletts

import (
	"net/http"
	"strings"

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

func (h *RouteHandler) CollectMethodRoutes(methods MethodIndex) {
	helpRoute := Route{
		Pattern: "/help",
		Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
			helper := make(map[string][]string)
			for name, method := range methods.Methods {
				helper["methods"] = append(helper["methods"], name)
				if method.Route == "" {
					helper["io_methods"] = append(helper["io_methods"], name)
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
			if method, ok := methods.Methods[name]; !ok {
				return cerrorf(RpcMethodNotFound, "The limit does not exist! %s", name).ResponseMessage()
			} else {
				return SuccessMsg(method)
			}
		},
	}
	h.Routes = append(h.Routes, []Route{subhelpRoute, helpRoute}...)
	for n, m := range methods.Methods {
		if m.Route != "" {
			name := n // for the sake of the closure
			method := m
			h.Routes = append(h.Routes, Route{
				Pattern: method.Route,
				Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
					//TODO get url params and merge w/ ivars
					req := &RequestMessage{
						Params: VarsInterface(mux.Vars(r)),
						Method: &name,
					}
					ctx := &RequestContext{
						HTTPWriter: &w,
						HTTPReader: r,
						Request:    req,
					}
					return ctx.Execute()
				},
			})
		}
	}
}

type HTTPHandler func(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage)

func (handle HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info("Request - %v", r)
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

// Reads the tile, dynamically determines enconding and content-type
func rawTileHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	//TODO wrap RequestContext and use method
	ivars := make(map[string]interface{})
	for k, v := range mux.Vars(r) {
		// cast xyz to float64
		if fv, err := atof(v); err == nil {
			ivars[k] = fv
		} else {
			ivars[k] = v
		}
	}
	resp := methods.Execute(nil)
	if resp.Result == nil {
		return
	} else if dojson := r.URL.Query().Get("json"); strings.ToLower(dojson) == "true" {
		msg = resp
		return
	}
	if tile, err := castTile(resp.Result); err != nil {
		errorlog(err)
		msg = cerrorf(http.StatusInternalServerError, "Internal Error casting tile contents").ResponseMessage()
	} else {
		//TODO roll sniff encoding into tile object?
		headers := formatEncoding[tile.SniffFormat()]
		for _, h := range headers {
			w.Header().Set(h.key, h.value)
		}
		w.Header().Set("Content-Length", sprintSizeOf(tile.Data))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(tile.Data)
	}
	return
}

// From http://www.jsonrpc.org/specification
// Content-Type: MUST be application/json.
// Content-Length: MUST contain the correct length according to the HTTP-specification.
// Accept: MUST be application/json.
func rpcHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
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
				HTTPWriter: &w,
				HTTPReader: r,
				Request:    req,
			}
			msg = ctx.Execute()
		}
	}
	return
}

func ioHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
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
	if cerr := c.listen(); cerr != nil {
		msg = cerr.ResponseMessage()
	} else {
		msg = SuccessMsg("WS connection closed succesfully")
	}
	//--> SUBSCRIBE on tile load -> {tileset, tileXYZ}
	//--> UNSUBSCRIBE on tile unload -> {tileset, tileXYZ}
	//--> LIST_SUBSCRIPTIONS
	//<-- {tileset, tile, data, type}
	return
}

type RequestContext struct {
	Request    *RequestMessage
	Connection *connection
	HTTPWriter *http.ResponseWriter
	HTTPReader *http.Request
	Params     MethodParams
}

func (ctx *RequestContext) Execute() (msg *ResponseMessage) {
	msg = methods.Execute(ctx)
	msg.Id = ctx.Request.Id
	return
}

func VarsInterface(vars map[string]string) map[string]interface{} {
	ivars := make(map[string]interface{})
	for k, v := range vars {
		ivars[k] = v
	}
	return ivars
}
