package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	ts "github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/mbtiles"
)

func GeoJsonToMbtiles(gjpath, mbtpath string, extent uint) {
	ts.TileSize = extent
	attrs := map[string]string{
		"name":        "test",
		"type":        "overlay",
		"version":     "1",
		"description": "some info here",
		"format":      "pbf.gz",
	}
	ts, err := mbtiles.InitTileset(mbtpath, attrs)
	if err != nil {
		panic(err)
	}
	var zoom uint = 13
	collection := readGeoJson(gjpath)
	tiles := splitFeatures(publishFeatureCollection(collection), zoom)
	for tile, features := range tiles {
		aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
		aLayer := aTile.NewLayer("denver", extent)
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
