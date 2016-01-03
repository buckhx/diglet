package transform

import (
	"encoding/json"
	"fmt"
	"github.com/buckhx/diglet/transform/tile_system"
	"github.com/deckarep/golang-set"
	"github.com/kpawlik/geojson"
	"io/ioutil"

	//"github.com/buckhx/diglet/transform/mvt"
	//"github.com/buckhx/diglet/transform/mvt/vector_tile"
)

// Split features up by their tile coordinates. This is intended to be done at the deepest desired zoom level
func splitFeatures(features <-chan *geojson.Feature, zoom uint) (tiles map[tile_system.Tile][]*geojson.Feature) {
	tiles = make(map[tile_system.Tile][]*geojson.Feature)
	for feature := range features {
		if feature.Type != "Feature" {
			continue
		}
		feature_tiles := mapset.NewSet()
		for _, point := range featurePoints(feature) {
			lat := float64(point[1])
			lon := float64(point[0])
			tile, _ := tile_system.CoordinateToTile(lat, lon, zoom)
			feature_tiles.Add(tile)
		}
		for t := range feature_tiles.Iter() {
			tile := t.(tile_system.Tile)
			tiles[tile] = append(tiles[tile], feature)
		}
	}
	return
}

// Flatten all the points of a feature into single list. This can hel in identifying which tiles are going to be
// created
func featurePoints(feature *geojson.Feature) (points []geojson.Coordinate) {
	// TODO: This sucks... I just want to switch on Coordinates.(type)
	igeom, err := feature.GetGeometry()
	check(err)
	switch geom := igeom.(type) {
	case *geojson.Point:
		coords := geom.Coordinates
		points = append(points, coords)
	case *geojson.LineString:
		coords := geom.Coordinates
		points = coords
	case *geojson.MultiPoint:
		coords := geom.Coordinates
		points = coords
	case *geojson.MultiLineString:
		coords := geom.Coordinates
		for _, line := range coords {
			for _, point := range line {
				points = append(points, point)
			}
		}
	case *geojson.Polygon:
		coords := geom.Coordinates
		for _, line := range coords {
			for _, point := range line {
				points = append(points, point)
			}
		}
	case *geojson.MultiPolygon:
		lines := geom.Coordinates
		for _, coords := range lines {
			for _, line := range coords {
				for _, point := range line {
					points = append(points, point)
				}
			}
		}
	default:
		panic("Invalid Coordinate Type in Feature") // + feature.String())
	}
	return
}

func readGeoJson(path string) (features *geojson.FeatureCollection) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		check(err)
	}
	if err := json.Unmarshal(file, &features); err != nil {
		check(err)
	}
	return features
}

func publishFeatureCollection(collection *geojson.FeatureCollection) (features chan *geojson.Feature) {
	features = make(chan *geojson.Feature, 10)
	go func() {
		defer close(features)
		for _, feature := range collection.Features {
			features <- feature
		}
	}()
	return
}
