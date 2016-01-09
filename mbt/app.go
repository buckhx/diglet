package mbt

import (
	"fmt"
	"github.com/buckhx/diglet/mbt/mvt"
)

func GeoJsonToMbtiles(gjpath, mbtpath string) {
	/*
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
	*/
	var zoom uint = 13
	collection := readGeoJson(gjpath)
	tiles := splitFeatures(publishFeatureCollection(collection), zoom)
	for tile, features := range tiles {
		aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
		layer := "denver"
		aTile.AddLayer(layer, 256)
		for _, feature := range features {
			aTile.AddFeature(layer, feature.ToMvtAdapter(zoom))
		}
		for _, layer := range aTile.GetTile().GetLayers() {
			for _, feature := range layer.GetFeatures() {
				fmt.Printf("%v", feature)
				geom := mvt.GeometryFromVectorTile(feature.Geometry)
				fmt.Printf("%v", geom.ToCommands())
			}
		}
		/*
			gz, err := aTile.GetTileGz()
			if err != nil {
				panic(err)
			}
			ts.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
		*/
	}
}
