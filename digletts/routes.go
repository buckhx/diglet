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
		Route{"/rpc", rpcHandler},
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
	xyz := make([]float64, 3)
	for i, v := range []string{vars["x"], vars["y"], vars["z"]} {
		iv, err := atoi(v)
		if err != nil {
			msg = ErrorMsg(http.StatusBadRequest, "Could not parse url tile coordinate param: "+v)
			return
		}
		xyz[i] = float64(iv)
		// json passes nums as float64 and we leverage the json-rpc implementation
		// we could make an XYZ struct that could be helpful (in mbtiles)
	}
	params := map[string]interface{}{
		"tileset": vars["ts"],
		"x":       xyz[0],
		"y":       xyz[1],
		"z":       xyz[2],
	}
	resp, rerr := methods.Execute(GetTile, params)
	if rerr != nil {
		msg = ErrorMsg(http.StatusBadRequest, rerr.Message)
		return
	}
	tile, err := assertTile(resp.Result)
	if err != nil {
		msg = ErrorMsg(http.StatusInternalServerError, "Internal Error asserting tile contents")
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

// From http://www.jsonrpc.org/specification
// Content-Type: MUST be application/json.
// Content-Length: MUST contain the correct length according to the HTTP-specification.
// Accept: MUST be application/json.
func rpcHandler(w http.ResponseWriter, r *http.Request) (msg *ResponseMessage) {
	switch {
	case r.Method != "POST":
		msg = ErrorMsg(http.StatusMethodNotAllowed, "Requires method: POST")
	case r.Header.Get("Content-Type") != "application/json":
		msg = ErrorMsg(http.StatusUnsupportedMediaType, "Requires Content-Type: application/json")
	case r.Header.Get("Accept") != "application/json":
		msg = ErrorMsg(http.StatusNotAcceptable, "Requires Accept: application/json")
	case r.Header.Get("Content-Length") == "":
		//TODO is it necessary to asset lenght is correct?
		msg = ErrorMsg(http.StatusLengthRequired, "Requires valid Content-Length")
	default:
		if req, rerr := ReadRequestMessage(r.Body); rerr != nil {
			msg = ErrorMsg(rerr.Code, rerr.Message)
		} else {
			msg = req.ExecuteMethod()
		}
	}
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
