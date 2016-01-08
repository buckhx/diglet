package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/mbtiles"
	"github.com/deckarep/golang-set"
	"os"
	"testing"
)

func TestSplitFeatures(t *testing.T) {
	var zoom uint = 13
	collection := readGeoJson("test_data/denver_features.geojson")
	features := publishFeatureCollection(collection)
	want := mapset.NewSetFromSlice([]interface{}{
		tile_system.Tile{X: 1707, Y: 3110, Z: 13},
		tile_system.Tile{X: 1706, Y: 3108, Z: 13},
		tile_system.Tile{X: 1706, Y: 3109, Z: 13},
	})
	got := mapset.NewSet()
	tiles := splitFeatures(features, zoom)
	for tile, _ := range tiles {
		got.Add(tile)
	}
	if !want.Equal(got) {
		t.Errorf("Did not get tiles %v -> %v", want, got)
	}
}

func TestWriteTiles(t *testing.T) {
	test_mbtiles := "test_data/test.mbtiles"
	defer os.Remove(test_mbtiles)
	attrs := map[string]string{
		"name":        "test",
		"type":        "overlay",
		"version":     "1",
		"description": "some info here",
		"format":      "pbf.gz",
	}
	ts, err := mbtiles.InitTileset(test_mbtiles, attrs)
	if err != nil {
		t.Error(err)
	}
	var zoom uint = 13
	collection := readGeoJson("test_data/denver_features.geojson")
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
			t.Error(err)
		}
		ts.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
	}
}
