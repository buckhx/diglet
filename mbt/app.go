package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/mbtiles"
)

func CreateTileset(mbtpath, desc string, extent uint) (ts *mbtiles.Tileset, err error) {
	tile_system.TileSize = extent
	attrs := map[string]string{
		"name":        util.SlugBase(mbtpath),
		"type":        "overlay",
		"version":     "1",
		"description": desc,
		"format":      "pbf.gz",
	}
	ts, err = mbtiles.InitTileset(mbtpath, attrs)
	return
}

func GeojsonTileset(ts *mbtiles.Tileset, gjpath string, zmin, zmax uint) {
	zoom := zmax
	collection := readGeoJson(gjpath)
	tiles := splitFeatures(publishFeatureCollection(collection), zoom)
	for tile, features := range tiles {
		aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
		aLayer := aTile.NewLayer("denver", tile_system.TileSize)
		for _, feature := range features {
			aFeature := feature.ToMvtAdapter(tile)
			aLayer.AddFeature(aFeature)
		}
		/*
			for _, layer := range aTile.GetTile().GetLayers() {
				for _, feature := range layer.GetFeatures() {
					fmt.Printf("%v\n", feature)
					geom := mvt.GeometryFromVt(*feature.Type, feature.Geometry)
					for _, cmd := range geom.ToCommands() {
						fmt.Printf("\t%v\n", cmd)
					}
				}
			}
		*/
		gz, err := aTile.GetTileGz()
		if err != nil {
			panic(err)
		}
		ts.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
	}
}
