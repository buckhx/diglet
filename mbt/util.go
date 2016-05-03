package mbt

import (
	"github.com/buckhx/diglet/geo"
	ts "github.com/buckhx/diglet/geo/tile_system"
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/deckarep/golang-set"
)

// CoverZoom sets the zoom level for flat coverings
// This is a global b/c it could go away if a dynamic cover is implemented
var CoverZoom = 15

// FeatureTiles returns a list of tiles that cover the feature at CoverZoom level
// Dups are not checked for, so they can exist
func FeatureTiles(f *geo.Feature) (tiles []ts.Tile) {
	for _, s := range f.Geometry {
		tiles = append(tiles, ShapeTiles(s)...)
	}
	return
}

// Shape tiles returns a list of tiles that cover a shape at CoverZoom level
func ShapeTiles(shp *geo.Shape) (tiles []ts.Tile) {
	bb := shp.BoundingBox()
	ne := bb.NorthEast().ToTile(CoverZoom)
	sw := bb.SouthWest().ToTile(CoverZoom)
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
