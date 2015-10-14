package digletts

import (
	"encoding/binary"
	"encoding/json"
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

type header struct {
	key, value string
}

var formatEncoding = map[mbtiles.Format][]header{
	mbtiles.PNG:     []header{header{"Content-Type", "image/png"}},
	mbtiles.JPG:     []header{header{"Content-Type", "image/jpeg"}},
	mbtiles.GIF:     []header{header{"Content-Type", "image/gif"}},
	mbtiles.WEBP:    []header{header{"Content-Type", "image/webp"}},
	mbtiles.PBF_GZ:  []header{header{"Content-Type", "application/x-protobuf"}, header{"Content-Encoding", "gzip"}},
	mbtiles.PBF_DF:  []header{header{"Content-Type", "application/x-protobuf"}, header{"Content-Encoding", "deflate"}},
	mbtiles.UNKNOWN: []header{header{"Content-Type", "application/octet-stream"}},
	mbtiles.EMPTY:   []header{header{"Content-Type", "application/octet-stream"}},
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
	headers := formatEncoding[tile.SniffFormat()]
	for _, h := range headers {
		w.Header().Set(h.key, h.value)
	}
	w.Header().Set("Content-Length", strconv.Itoa(binary.Size(tile.Data)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(tile.Data)
}

func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	attrs, err := json.Marshal(ts.Metadata().Attributes())
	if check(w, err) == true {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(attrs)
}

func (s *Server) Start() (err error) {
	log.Println("Starting server...")

	r := mux.NewRouter()
	r.HandleFunc("/tileset/{z}/{x}/{y}", TileHandler)
	r.HandleFunc("/tileset/metadata", MetadataHandler)
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
