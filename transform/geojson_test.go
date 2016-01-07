package transform

import (
	"github.com/buckhx/diglet/transform/mvt"
	"github.com/buckhx/diglet/transform/tile_system"
	"github.com/deckarep/golang-set"
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
		vtile := aTile.GetTile()
		t.Errorf("%s", vtile.String())
	}
}
