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
		//Route{"/io/{ts}", ioHandler},
		Route{"/{ts}/{z}/{x}/{y}", tileHandler},
		Route{"/{ts}", metadataHandler},
		Route{"/", listHandler},
	}}
	/*
		go func() {
			for event := range tilesets.Events {
				tilehub.broadcast <- event
				info("Tileset Change - %s", event.String())
			}
		}()
	*/
	return
}

// Reads the tile, dynamically determines enconding and content-type
func tileHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
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
func metadataHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	//TODO if there's a json field, try to deserialze that
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := tilesets.Tilesets[slug]; ok {
		msg = SuccessMsg(ts.Metadata().Attributes())
	} else {
		msg = ErrorMsg(http.StatusBadRequest, fmt.Sprintf("No tileset named %q", slug))
	}
	return
}

// List the tilesets available on the server
func listHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	tss := make(map[string]map[string]string)
	for name, ts := range tilesets.Tilesets {
		tss[name] = ts.Metadata().Attributes()
	}
	msg = SuccessMsg(tss)
	return
}

/*
func ioHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	if r.Method != "GET" {
		msg = ErrorMsg(http.StatusMethodNotAllowed, "Only GET can be upgraded")
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		warn(err)
		msg = ErrorMsg(http.StatusInternalServerError, "Request can't be upgraded")
		return
	}
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := tilesets.Tilesets[slug]; !ok {
		msg = ErrorMsg(http.StatusBadRequest, fmt.Sprintf("No tileset named %q", slug))
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	c.readPump()
}
*/
