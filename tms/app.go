// Package diglet/tms is an HTTP Tile Server that also support JSON-RPC & WebSocket requests. Tile subscriptions
// are also available to support real-time map applications with large feature sets.
package tms

import (
	dig "github.com/buckhx/diglet/burrow"
	"github.com/buckhx/diglet/resources"
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

func MBTServer(mbtPath string, port string) (app *dig.App, err error) {
	tilesets, err = ReadTilesets(mbtPath)
	if err != nil {
		return
	}
	info("Serving tiles from %s", mbtPath)
	hub = NewHub(tilesets)
	go hub.listen()
	app = dig.NewApp("Diglet")
	app.Port = port
	app.Prefix = "/tileset"
	app.Methods = []dig.Method{
		{
			Name:  "gallery",
			Route: "/gallery/{tileset}",
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to query for metadata"},
			},
			Handler: func(ctx *dig.RequestContext) (t interface{}, err *dig.CodedError) {
				template := resources.Static_html()
				static, errp := ctx.Render(template)
				warn(errp, "params?")
				//panic(errp)
				w := ctx.HTTPWriter
				w.Write([]byte(static))
				return
			},
			Help: "A simple tile viewer gallery app.\n" +
				"?lat={}&lon={}&zoom={} can be included to zoom tolocation",
		},
		{
			Name: GetTile,
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to read from"},
				"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
				"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
				"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
			},
			Handler: getTileHandler,
			Help:    "Retrieve a tile, the response's data field will be binary of the contents",
		},
		{
			Name:   ListTilesets,
			Route:  "/",
			Params: dig.MethodParams{},
			Handler: func(ctx *dig.RequestContext) (tile interface{}, err *dig.CodedError) {
				dict := make(map[string]map[string]string)
				for name, ts := range tilesets.Tilesets {
					dict[name] = ts.Metadata().Attributes()
				}
				return dict, nil
			},
			Help: "List all of the tilesets available, including their metadata",
		},
		{
			Name:  GetTileset,
			Route: "/{tileset}",
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to query for metadata"},
			},
			Handler: func(ctx *dig.RequestContext) (attrs interface{}, err *dig.CodedError) {
				params := ctx.Params
				slug := params["tileset"].GetString()
				if ts, ok := tilesets.Tilesets[slug]; ok {
					attrs = ts.Metadata().Attributes()
				} else {
					err = dig.Cerrorf(dig.RpcInvalidRequest, "No tileset named %s", slug)
				}
				return
			},
			Help: "Query for the tilesets metadata, all values are string representations",
		},
		{
			Name: SubscribeTile,
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
				"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
				"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
				"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
			},
			Handler: func(ctx *dig.RequestContext) (res interface{}, err *dig.CodedError) {
				params := ctx.Params
				x := params["x"].GetInt()
				y := params["y"].GetInt()
				z := params["z"].GetInt()
				slug := params["tileset"].GetString()
				if _, ok := tilesets.Tilesets[slug]; !ok {
					err = dig.Cerrorf(dig.RpcInvalidRequest, "Cannot find tileset %s", slug)
				} else {
					xyz := TileXYZ{Tileset: slug, X: x, Y: y, Z: z}
					if e := hub.bindTile(ctx, xyz); err != nil {
						err = dig.Cerrorf(dig.RpcInvalidRequest, e.Error())
					} else {
						// might need to make this a notifucation instead
						// -> no msg.Id
						res = sprintf("Subscribed to tile %s", xyz)
					}
				}
				return
			},
			Help: "Subscribe to changes on a specific tile, changes will be pushd with the same request id",
		},
		{
			Name: UnsubscribeTile,
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
				"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
				"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
				"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
			},
			Handler: func(ctx *dig.RequestContext) (res interface{}, err *dig.CodedError) {
				params := ctx.Params
				x := params["x"].GetInt()
				y := params["y"].GetInt()
				z := params["z"].GetInt()
				slug := params["tileset"].GetString()
				if _, ok := tilesets.Tilesets[slug]; !ok {
					err = dig.Cerrorf(dig.RpcInvalidRequest, "Cannot find tileset %s", slug)
				} else {
					xyz := TileXYZ{Tileset: slug, X: x, Y: y, Z: z}
					hub.unbindTile(ctx, xyz)
					res = sprintf("Unsubscribed from tile %s", xyz)
				}
				return
			},
			Help: "Unsubscribe from a tile",
		},
		{
			Name:  GetRawTile,
			Route: "/{tileset}/{z}/{x}/{y}",
			Params: dig.MethodParams{
				"tileset": {Validator: assertString, Help: "Tileset to subscribe to"},
				"x":       {Validator: assertNumber, Help: "E/W Coordinate"},
				"y":       {Validator: assertNumber, Help: "N/S Cooredinate"},
				"z":       {Validator: assertNumber, Help: "Zoom level Coordinate"},
			},
			Handler: func(ctx *dig.RequestContext) (otile interface{}, err *dig.CodedError) {
				itile, err := getTileHandler(ctx)
				if err != nil {
					return
				}
				r := ctx.HTTPReader
				w := ctx.HTTPWriter
				if tile, terr := castTile(itile); err != nil {
					errorlog(terr)
					terr = dig.Cerrorf(500, "Internal Error casting tile contents")
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
	}
	return
}

// This is pulled out so that get_tile & get_raw_tile endpoitn can use the same code
func getTileHandler(ctx *dig.RequestContext) (tile interface{}, err *dig.CodedError) {
	params := ctx.Params
	x := params["x"].GetInt()
	y := params["y"].GetInt()
	z := params["z"].GetInt()
	slug := params["tileset"].GetString()
	xyz := TileXYZ{Tileset: slug, X: x, Y: y, Z: z}
	tile, tserr := tilesets.Read(xyz)
	if tserr != nil {
		err = dig.Cerrorf(500, tserr.Error())
	}
	return
}
