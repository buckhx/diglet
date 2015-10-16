package digletts

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buckhx/mbtiles"
	"github.com/gorilla/mux"
)

var bag *TilesetBag

func TilesetRoutes(prefix, mbtPath string) (r *RouteHandler) {
	bag = ReadTilesets(mbtPath)
	r = &RouteHandler{prefix, []Route{
		Route{"/{ts}/{z}/{x}/{y}", TileHandler},
		Route{"/{ts}", MetadataHandler},
		Route{"/", ListHandler},
	}}
	return
}

func TileHandler(w http.ResponseWriter, r *http.Request) (content []byte, err error) {
	vars := mux.Vars(r)
	tile, err := bag.tileFromVars(vars)
	if err != nil {
		return
	}
	headers := formatEncoding[tile.SniffFormat()]
	for _, h := range headers {
		w.Header().Set(h.key, h.value)
	}
	w.Header().Set("Content-Length", strconv.Itoa(binary.Size(tile.Data)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(tile.Data)
	return
}

func MetadataHandler(w http.ResponseWriter, r *http.Request) (content []byte, err error) {
	//TODO if there's a json field, try to deserialze that
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := bag.Tilesets[slug]; ok {
		content, err = json.Marshal(ts.Metadata().Attributes())
	} else {
		err = fmt.Errorf("No tileset named %q", slug)
	}
	return
}

func ListHandler(w http.ResponseWriter, r *http.Request) (content []byte, err error) {
	//TODO include refresh and metdata flag
	names := make([]string, 0, len(bag.Tilesets))
	for name := range bag.Tilesets {
		names = append(names, name)
	}
	content, err = json.Marshal(names)
	return
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

type TilesetBag struct {
	Path     string
	Tilesets map[string]*mbtiles.Tileset
}

func ReadTilesets(dir string) (bag *TilesetBag) {
	bag = &TilesetBag{dir, make(map[string]*mbtiles.Tileset)}
	mbtPaths, err := filepath.Glob(filepath.Join(dir, "*.mbtiles"))
	check(err)
	for _, path := range mbtPaths {
		ts, err := mbtiles.ReadTileset(path)
		if err != nil {
			warn(err, "skipping "+path)
			continue
		}
		name := cleanTilesetName(path)
		if _, exists := bag.Tilesets[name]; exists {
			check(fmt.Errorf("Multiple tilesets with slug %q like %q", name, path))
		}
		bag.Tilesets[name] = ts
	}
	return
}

func (bag *TilesetBag) tileFromVars(vars map[string]string) (tile *mbtiles.Tile, err error) {
	slug := vars["ts"]
	x, err := strconv.Atoi(vars["x"])
	y, err := strconv.Atoi(vars["y"])
	z, err := strconv.Atoi(vars["z"])
	if ts, ok := bag.Tilesets[slug]; ok && err == nil {
		tile, err = ts.ReadSlippyTile(x, y, z)
	} else {
		err = fmt.Errorf("No tileset with slug %q", slug)
	}
	return
}

func cleanTilesetName(path string) (slug string) {
	f := filepath.Base(path)
	f = strings.TrimSuffix(f, filepath.Ext(f))
	slug = slugged(f)
	return
}
