package mbt

import (
	ts "github.com/buckhx/diglet/mbt/tile_system"
	"github.com/deckarep/golang-set"
)

// Split features up by their tile coordinates. This is intended to be done at the deepest desired zoom level
// If a feature has any point in a tile, it will bind to that tile. A feature can be in multiple tiles
func splitFeatures(features <-chan *Feature, zoom uint) (tiles map[ts.Tile][]*Feature) {
	tiles = make(map[ts.Tile][]*Feature)
	for feature := range features {
		feature_tiles := mapset.NewSet()
		for _, shape := range feature.Geometry {
			for _, c := range shape.Coordinates {
				tile, _ := ts.CoordinateToTile(c.Lat, c.Lon, zoom)
				feature_tiles.Add(tile)
			}
		}
		for t := range feature_tiles.Iter() {
			tile := t.(ts.Tile)
			tiles[tile] = append(tiles[tile], feature)
		}
	}
	return
}
