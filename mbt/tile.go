package mbt

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/mbt/mvt"
	ts "github.com/buckhx/diglet/mbt/tile_system"
	"github.com/deckarep/golang-set"
)

// Split features up by their tile coordinates. This is intended to be done at the deepest desired zoom level
// If a feature has any point in a tile, it will bind to that tile. A feature can be in multiple tiles
func splitFeatures(features <-chan *geo.Feature, zoom uint) (tiles map[ts.Tile][]*geo.Feature) {
	tiles = make(map[ts.Tile][]*geo.Feature)
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

// MvtAdapter populates an mvt feature from a diglet geo feature
func MvtAdapter(f *geo.Feature, t ts.Tile) (a *mvt.Feature) {
	g := f.Type
	if g == geo.LineFeature {
		g = "linestring"
	}
	a = mvt.NewFeatureAdapter(f.GetUint64ID(), g, f.Properties)
	shps := tiledShapes(f, t)
	a.AddShape(shps...)
	return
}

func tiledShapes(f *geo.Feature, t ts.Tile) (shps []*mvt.Shape) {
	shps = make([]*mvt.Shape, len(f.Geometry))
	for i, s := range f.Geometry {
		shp := tiledShape(s, t)
		shps[i] = shp
	}
	return
}

func tiledShape(s *geo.Shape, t ts.Tile) (shp *mvt.Shape) {
	shp = mvt.MakeShape(len(s.Coordinates))
	for i, c := range s.Coordinates {
		pixel := ts.ClippedCoords(c.Lat, c.Lon).ToPixel(t.Z)
		x := int(pixel.X - t.ToPixel().X)
		y := int(pixel.Y - t.ToPixel().Y)
		point := mvt.Point{X: x, Y: y}
		//TODO: clipping
		shp.Insert(i, point)
	}
	return
}
