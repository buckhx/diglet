// Package digletss is a diglet tile server
package digletts

import (
	"github.com/gorilla/mux"
)

var hub *IoHub
var tilesets *TilesetIndex

const (
	GetTile         string = "get_tile"
	GetRawTile      string = "get_raw_tile"
	GetTileset      string = "get_tileset"
	ListTilesets    string = "list_tilesets"
	SubscribeTile   string = "subscribe_tile"
	UnsubscribeTile string = "unsubscribe_tile"
)

var methods = MethodIndex{Methods: map[string]Method{
	GetTile: Method{
		Name: GetTile,
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to read from"},
			"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
			"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
			"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
		},
		Handler: getTileHandler,
		Help:    "Retrieve a tile, the response's data field will be binary of the contents",
	},
	ListTilesets: Method{
		Name:   ListTilesets,
		Route:  "/",
		Params: MethodParams{},
		Handler: func(ctx *RequestContext) (tile interface{}, err *CodedError) {
			dict := make(map[string]map[string]string)
			for name, ts := range tilesets.Tilesets {
				dict[name] = ts.Metadata().Attributes()
			}
			return dict, nil
		},
		Help: "List all of the tilesets available, including their metadata",
	},
	GetTileset: Method{
		Name:  GetTileset,
		Route: "/{tileset}",
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to query for metadata"},
		},
		Handler: func(ctx *RequestContext) (attrs interface{}, err *CodedError) {
			params := ctx.Params
			slug := params["tileset"].GetString()
			if ts, ok := tilesets.Tilesets[slug]; ok {
				attrs = ts.Metadata().Attributes()
			} else {
				err = cerrorf(RpcInvalidRequest, "No tileset named %s", slug)
			}
			return
		},
		Help: "Query for the tilesets metadata, all values are string representations",
	},
	SubscribeTile: Method{
		Name: SubscribeTile,
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
			"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
			"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
			"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
		},
		Handler: func(ctx *RequestContext) (tile interface{}, err *CodedError) {
			params := ctx.Params
			x := params["x"].GetInt()
			y := params["y"].GetInt()
			z := params["z"].GetInt()
			slug := params["tileset"].GetString()
			if _, ok := tilesets.Tilesets[slug]; !ok {
				err = cerrorf(RpcInvalidRequest, "Cannot find tileset %s", slug)
			} else {
				xyz := TileXYZ{Tileset: slug, X: x, Y: y, Z: z}
				ctx.Connection.bindTile(xyz, ctx.Request.Id)
				ctx.Connection.notify("Subscribed to tile %s", xyz)
			}
			return
		},
		Help: "Subscribe to changes on a specific tile, changes will be pushd with the same request id",
	},
	UnsubscribeTile: Method{
		Name: UnsubscribeTile,
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
			"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
			"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
			"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
		},
		Handler: func(ctx *RequestContext) (v interface{}, err *CodedError) {
			params := ctx.Params
			x := params["x"].GetInt()
			y := params["y"].GetInt()
			z := params["z"].GetInt()
			slug := params["tileset"].GetString()
			if _, ok := tilesets.Tilesets[slug]; !ok {
				err = cerrorf(RpcInvalidRequest, "Cannot find tileset %s", slug)
			} else {
				xyz := TileXYZ{Tileset: slug, X: x, Y: y, Z: z}
				ctx.Connection.unbindTile(xyz)
				ctx.Connection.notify("Unsubscribed from tile %s", xyz)
			}
			return
		},
		Help: "Unsubscribe from a tile",
	},
	GetRawTile: Method{
		Name:  GetRawTile,
		Route: "/{tileset}/{z}/{x}/{y}",
		Params: MethodParams{
			"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
			"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
			"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
			"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
		},
		Handler: func(ctx *RequestContext) (otile interface{}, err *CodedError) {
			itile, err := getTileHandler(ctx)
			if err != nil {
				return
			}
			r := ctx.HTTPReader
			w := ctx.HTTPWriter
			if tile, terr := castTile(itile); err != nil {
				errorlog(terr)
				terr = cerrorf(500, "Internal Error casting tile contents")
			} else {
				if dojson := r.URL.Query().Get("json"); toLower(dojson) == "true" {
					otile = tile
					return
				}
				//TODO roll sniff encoding into tile object?
				headers := formatEncoding[tile.SniffFormat()]
				for _, h := range headers {
					w.Header().Set(h.key, h.value)
				}
				w.Header().Set("Content-Length", sprintSizeOf(tile.Data))
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Write(tile.Data)
			}
			return

		},
		Help: "Gets a tile and only writes it's raw contents. Used for hosting static tiles.",
	},
}}

// This is pulled out so that get_tile & get_raw_tile endpoitn can use the same code
func getTileHandler(ctx *RequestContext) (tile interface{}, err *CodedError) {
	params := ctx.Params
	x := params["x"].GetInt()
	y := params["y"].GetInt()
	z := params["z"].GetInt()
	slug := params["tileset"].GetString()
	if ts, ok := tilesets.Tilesets[slug]; !ok {
		err = cerrorf(RpcInvalidRequest, "Cannot find tileset %s", slug)
	} else {
		var tserr error
		if tile, tserr = ts.ReadSlippyTile(x, y, z); tserr != nil {
			err = cerrorf(RpcInvalidRequest, tserr.Error())
		}
	}
	return
}

func MBTServer(mbtPath, port string) (s *Server, err error) {
	port = ":" + port
	r := mux.NewRouter()
	r.StrictSlash(true)
	tilesets = ReadTilesets(mbtPath)
	hub := NewHub(tilesets)
	go hub.publish(tilesets.Events)
	routes := &RouteHandler{"/tileset", []Route{
		Route{"/io", ioHandler},
		Route{"/rpc", rpcHandler},
	}}
	routes.CollectMethodRoutes(methods)
	routes.Subrouter(r)
	s = &Server{
		Router:  r,
		DataDir: mbtPath,
		Port:    port,
	}
	return
}
