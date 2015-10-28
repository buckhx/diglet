package digletts

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var tilesets *TilesetIndex

func TilesetRoutes(prefix, mbtPath string) (r *RouteHandler) {
	tilesets = ReadTilesets(mbtPath)
	r = &RouteHandler{prefix, []Route{
		Route{"/{ts}/{z}/{x}/{y}", TileHandler},
		Route{"/{ts}", MetadataHandler},
		Route{"/", ListHandler},
	}}
	go func() {
		for event := range tilesets.Events {
			info("Tileset Change - %s", event.String())
		}
	}()
	return
}

// Reads the tile, dynamically determines enconding and content-type
func TileHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	vars := mux.Vars(r)
	tile, err := tilesets.tileFromVars(vars)
	if err != nil {
		return
	}
	headers := formatEncoding[tile.SniffFormat()]
	for _, h := range headers {
		w.Header().Set(h.key, h.value)
	}
	w.Header().Set("Content-Length", sprintSizeOf(tile.Data))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(tile.Data)
	return
}

// Get the metadatadata map from the tileset
func MetadataHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	//TODO if there's a json field, try to deserialze that
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := tilesets.Tilesets[slug]; ok {
		response = Success(ts.Metadata().Attributes())
	} else {
		response = Error(http.StatusBadRequest, fmt.Sprintf("No tileset named %q", slug))
	}
	return
}

// List the tilesets available on the server
func ListHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	tss := make(map[string]map[string]string)
	for name, ts := range tilesets.Tilesets {
		tss[name] = ts.Metadata().Attributes()
	}
	response = Success(tss)
	return
}
