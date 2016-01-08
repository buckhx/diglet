package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/mbtiles"
)

func GeoJsonToMbtiles(gjpath, mbtpath string) {
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
		layer := "denver"
		aTile.AddLayer(layer, 256)
		for _, feature := range features {
			var fid *uint64
			if feature.Id != nil {
				val := uint64(feature.Id.(float64))
				fid = &val
			} else {
				fid = nil
			}
			geom, err := feature.GetGeometry()
			check(err)
			aFeature := mvt.NewFeatureAdapter(fid, geom.GetType())
			shapes := featureShapes(feature, zoom)
			aFeature.AddShapes(shapes)
			aTile.AddFeature(layer, aFeature)
			//properties := featureValues(feature)
		}
		/*
			for _, layer := range aTile.GetTile().GetLayers() {
				for _, feature := range layer.GetFeatures() {
					t.Errorf("%v", feature)
					geom := mvt.GeometryFromVectorTile(feature.Geometry)
					t.Errorf("%v", geom.ToCommands())
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
