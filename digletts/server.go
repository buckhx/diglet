package digletts

import (
	"log"
	"net/http"

	"github.com/buckhx/mbtiles"
)

type Server struct {
	TileData, Port string
}

func MBTServer(mbt_path, port string) (s *Server, err error) {
	port = ":" + port
	ts, err = mbtiles.ReadTileset(mbt_path)
	s = &Server{mbt_path, port}
	return
}

func check(w http.ResponseWriter, err error) (caught bool) {
	caught = false
	if err != nil {
		caught = true
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func checks(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

func (s *Server) Start() (err error) {
	log.Println("Starting server...")

	r := BuildRouter()
	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(static)
	http.Handle("/", r)

	log.Printf("Now serving tiles from %s on port %s\n", s.TileData, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
	return
}
