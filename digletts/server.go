// Package digletss is a diglet tile server
package digletts

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	DataDir, Port string
	Router        *mux.Router
}

func MBTServer(dataPath, port string) (s *Server, err error) {
	port = ":" + port
	r := mux.NewRouter()
	r.StrictSlash(true)
	_ = TilesetRoutes("/tileset", dataPath).Subrouter(r)
	s = &Server{
		Router:  r,
		DataDir: dataPath,
		Port:    port,
	}
	return
}

func (s *Server) Start() (err error) {
	info("Starting server...")

	s.mountStatic()
	http.Handle("/", s.Router)

	info("Now serving tiles from %s on port %s", s.DataDir, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	check(err)
	return
}

func (s *Server) mountStatic() {
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	s.Router.PathPrefix("/static/").Handler(static)
}

type Handler func(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage)

func (handle Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

type Route struct {
	Pattern string
	Handler Handler
}

type RouteHandler struct {
	Prefix string
	Routes []Route
}

func (rh *RouteHandler) Subrouter(r *mux.Router) (subrouter *mux.Router) {
	subrouter = r.PathPrefix(rh.Prefix).Subrouter()
	for _, route := range rh.Routes {
		subrouter.Handle(route.Pattern, route.Handler)
	}
	return
}
