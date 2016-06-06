package mbt

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/mbt/mvt"
	ts "github.com/buckhx/tiles"
	"github.com/deckarep/golang-set"
)

// FeatureTiles returns a list of tiles that cover the feature at the given zoom level
// Dups are not checked for, so they can exist
func FeatureTiles(f *geo.Feature, zoom int) (tiles []ts.Tile) {
	for _, s := range f.Geometry {
		tiles = append(tiles, ShapeTiles(s, zoom)...)
	}
	return
}

// Shape tiles returns a list of tiles that cover a shape at the given zoom level
func ShapeTiles(shp *geo.Shape, zoom int) (tiles []ts.Tile) {
	bb := shp.BoundingBox()
	ne := bb.NorthEast().ToTile(zoom)
	sw := bb.SouthWest().ToTile(zoom)
	cur := sw
	for x := sw.X; x <= ne.X; x++ {
		for y := sw.Y; y >= ne.Y; y-- { //origin is NW
			cur.X, cur.Y = x, y
			tiles = append(tiles, cur)
		}
	}
	return
}

// Split features up by their tile coordinates. This is intended to be done at the deepest desired zoom level
// If a feature has any point in a tile, it will bind to that tile. A feature can be in multiple tiles
func splitFeatures(features <-chan *geo.Feature, zoom int) (tiles map[ts.Tile][]*geo.Feature) {
	tiles = make(map[ts.Tile][]*geo.Feature)
	for feature := range features {
		feature_tiles := mapset.NewSet()
		for _, shape := range feature.Geometry {
			for _, c := range shape.Coordinates {
				tile := ts.FromCoordinate(c.Lat, c.Lon, zoom)
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
	var id uint64
	switch v := f.ID.(type) {
	case uint, uint32, uint64:
		id = v.(uint64)
	case int, int32, int64:
		id = uint64(v.(int64))
	default:
		// stay nil
	}
	a = mvt.NewFeatureAdapter(&id, g, f.Properties)
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
