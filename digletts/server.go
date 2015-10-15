package digletts

import (
	"log"
	"net/http"

	"github.com/buckhx/mbtiles"
	"github.com/gorilla/mux"
)

type Server struct {
	TileData, Port string
	Router         *mux.Router
}

//TODO tile provider interface
var ts *mbtiles.Tileset

func MBTServer(mbt_path, port string) (s *Server, err error) {
	port = ":" + port
	ts, err = mbtiles.ReadTileset(mbt_path)
	s = &Server{mbt_path, port, mux.NewRouter()}
	return
}

func (s *Server) Start() (err error) {
	log.Println("Starting server...")

	TilesetRoutes("/tileset").Subrouter(s.Router)
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	s.Router.PathPrefix("/static/").Handler(static)
	http.Handle("/", s.Router)

	log.Printf("Now serving tiles from %s on port %s\n", s.TileData, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
	return
}

type Handler func(w http.ResponseWriter, r *http.Request)

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
		subrouter.HandleFunc(route.Pattern, route.Handler)
	}
	return
}
