package digletts

import (
	"net/http"

	"github.com/gorilla/mux"
)

var tilesets *TilesetIndex

func TilesetRoutes(prefix, mbtPath string) (r *RouteHandler) {
	tilesets = ReadTilesets(mbtPath)
	r = &RouteHandler{prefix, []Route{
		//Route{"/io", ioHandler},
		Route{"/rpc", rpcHandler},
		Route{"/{tileset}/{z}/{x}/{y}", rawTileHandler},
	}}
	r.CollectMethodRoutes(methods)
	/*
		go func() {
			for event := range tilesets.Events {
				tilehub.broadcast <- event
				info("Tileset Change - %s", event.String())
			}
		}()
	*/
	return
}

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
			for name, _ := range methods.Methods {
				helper["methods"] = append(helper["methods"], name)
			}
			helper["info"] = append(helper["info"], "Use help/{method} for method help")
			return SuccessMsg(helper)
		},
	}
	subhelpRoute := Route{
		Pattern: "/help/{method}",
		Handler: func(w http.ResponseWriter, r *http.Request) *ResponseMessage {
			name := mux.Vars(r)["method"]
			if method, ok := methods.Methods[name]; !ok {
				return cerrorf(RpcMethodNotFound, "The method does not exist! %s", name).ResponseMessage()
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
					ivars := make(map[string]interface{})
					for k, v := range mux.Vars(r) {
						ivars[k] = v
					}
					//TODO get url params and merge w/ ivars
					return methods.Execute(name, ivars)
				},
			})
		}
	}
}

// Reads the tile, dynamically determines enconding and content-type
func rawTileHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	ivars := make(map[string]interface{})
	for k, v := range mux.Vars(r) {
		// cast xyz to float64
		if fv, err := atof(v); err == nil {
			ivars[k] = fv
		} else {
			ivars[k] = v
		}
	}
	resp := methods.Execute(GetTile, ivars)
	if resp.Result == nil {
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
			msg = req.ExecuteMethod()
		}
	}
	return
}

/*
func ioHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	if r.Method != "GET" {
		msg = cerrorf(http.StatusMethodNotAllowed, "Only GET can be upgraded").ResponseMessage()
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorlog(err)
		msg = cerrorf(http.StatusInternalServerError, "Request can't be upgraded").ResponseMessage()
		return
	}
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := tilesets.Tilesets[slug]; !ok {
		msg = cerrorf(http.StatusBadRequest, fmt.Sprintf("No tileset named %q", slug)).ResponseMessage()
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	c.readPump()
}
*/
