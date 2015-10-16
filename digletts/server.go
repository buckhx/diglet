package digletts

import (
	"log"
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
	_ = TilesetRoutes("/tileset", dataPath).Subrouter(r)
	s = &Server{
		Router:  r,
		DataDir: dataPath,
		Port:    port,
	}
	return
}

func (s *Server) Start() (err error) {
	log.Println("Starting server...")

	s.mountStatic()
	http.Handle("/", s.Router)

	log.Printf("Now serving tiles from %s on port %s\n", s.DataDir, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	check(err)
	return
}

func (s *Server) mountStatic() {
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	s.Router.PathPrefix("/static/").Handler(static)
}

type Handler func(w http.ResponseWriter, r *http.Request) (content []byte, err error)

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if content, err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		if content != nil {
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
