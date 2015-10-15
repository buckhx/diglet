package digletts

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/buckhx/mbtiles"
	"github.com/gorilla/mux"
)

func TilesetRoutes(prefix string) (r *RouteHandler) {
	r = &RouteHandler{prefix, []Route{
		Route{"/{z}/{x}/{y}", TileHandler},
		Route{"/metadata", MetadataHandler},
		//Route{"/{db}/{z}/{x}/{y}", TileHandler},
		//Route{"/{db}/metadata", MetadataHandler},
		//Route{"/", ListHandler},
	}}
	return
}

func TileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	vars := mux.Vars(r)
	tile, _ := tileFromVars(vars)
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
	WriteJson(w, attrs)
}

/*
func ListHandler(w http.ResponseWriter, r *http.Request) {
	names = make([]string, 0, len(dbs))
	for name := dbs {
		names = append(names, name)
	}
	output, err := json.Marshal(ts.Metadata().Attributes())
	if check(w, err) == true {
		return
	}
	WriteJson(w, output)
}
*/

func WriteJson(w http.ResponseWriter, output []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

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

func tileFromVars(vars map[string]string) (tile *mbtiles.Tile, err error) {
	// TODO actually handle these
	x, err := strconv.Atoi(vars["x"])
	y, err := strconv.Atoi(vars["y"])
	z, err := strconv.Atoi(vars["z"])
	tile, err = ts.ReadSlippyTile(x, y, z)
	return
}

/*
type TilesetProvider struct {
	Path string
	Tilesets map[string]*Tileset
}

func readDBs(path string) map[]{

}
*/
