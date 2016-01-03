package transform

import (
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
		vtile := mvt.Tile{X, Y, Z}
		for _, feature := range features {
			properties :=

				vtile.AddFeature()
		}
	}
}
