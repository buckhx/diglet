package diglib

import (
	//"encoding/json"
	//"encoding/base64"
	"log"
	"net/http"
	"strconv"

	"github.com/buckhx/mbtiles"
	"github.com/gorilla/mux"
)

type Server struct {
	TileData, Port string
}

//TODO tile provider interface
var ts *mbtiles.Tileset

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

func TileFromVars(vars map[string]string) (tile *mbtiles.Tile, err error) {
	// TODO actually handle these
	x, err := strconv.Atoi(vars["x"])
	y, err := strconv.Atoi(vars["y"])
	z, err := strconv.Atoi(vars["z"])
	tile, err = ts.ReadSlippyTile(x, y, z)
	return
}

func TileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	vars := mux.Vars(r)
	tile, _ := TileFromVars(vars)
	/*
		t, err := json.Marshal(tile)
		if check(w, err) == true {
			return
		}
		w.Header().Set("Content-Type", "application/json")
	*/
	//img := base64.URLEncoding.EncodeToString(tile.Data)
	w.Header().Set("Content-Type", "image/png")
	w.Write(tile.Data)
}

func (s *Server) Start() (err error) {
	log.Println("Starting server...")

	r := mux.NewRouter()
	r.HandleFunc("/tile/{z}/{x}/{y}", TileHandler)
	http.Handle("/", r)

	log.Printf("Now serving tiles from %s on port %s\n", s.TileData, s.Port)
	err = http.ListenAndServe(s.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
	return
}
