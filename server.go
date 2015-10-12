package diglet

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

func NewServer(data, port string) *Server {
	return &Server{data, port}
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

var ts *mbtiles.Tileset

func TileFromVars(vars map[string]string) (tile *mbtiles.Tile, err error) {
	x, xerr := strconv.Atoi(vars["x"])
	y, xerr := strconv.Atoi(vars["y"])
	z, xerr := strconv.Atoi(vars["z"])
	tile = ts.ReadSlippyTile(x, y, z)
	err = xerr
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

func (s *Server) Start() {
	log.Println("Starting server...")
	ts = mbtiles.ReadTileset(s.TileData)

	r := mux.NewRouter()
	r.HandleFunc("/tile/{z}/{x}/{y}", TileHandler)
	http.Handle("/", r)

	log.Println("Listening...")
	err := http.ListenAndServe(s.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
