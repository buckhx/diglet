package mvt

import (
	vt "github.com/buckhx/diglet/transform/mvt/vector_tile"
	"strings"
)

type TileAdapter struct {
	x, y, z      uint
	tile         *vt.Tile
	layers       map[string]*vt.Tile_Layer
	keyIndexes   map[string]uint
	valueIndexes map[*vt.Tile_Value]uint
}

func NewTileAdapter(x, y, z uint) *TileAdapter {
	return &TileAdapter{
		x:    x,
		y:    y,
		z:    z,
		tile: &vt.Tile{},
	}
}

func (t *TileAdapter) AddLayer(name string, extent uint) {
	layer := &vt.Tile_Layer{
		Version:  2,
		Name:     name,
		Extent:   extent,
		Features: []*vt.Tile_Feature{},
		Keys:     []string{},
		Values:   []*vt.Tile_Value{},
	}
	t.layers[name] = layer
	t.tile.Layers = append(a.tile.Layers, layer)
}

func AddFeature(layer string, feature *FeatureAdapter) {

}

type FeatureAdapter struct {
	feature *vt.Tile_Feature
}

func NewFeatureAdapter(id uint) *FeatureAdapter {
	var gtype vt.Tile_GeomType
	switch strings.ToLower(geometry_type) {
	case "point", "multipoint":
		gtype = vt.Tile_POINT
	case "linestring", "multilinestring":
		gtype = vt.Tile_LINESTRING
	case "polygon", "multiPolygon":
		gtype = vt.Tile_POLYGON
	default:
		gtype = vt.Tile_UNKNOWN
	}
	feature := &vt.Tile_Feature{
		Id:       id,
		Tags:     []uint{},
		Type:     *gtype,
		Geometry: []uint{},
	}
	return &FeatureAdapter{feature: feature}
}

func (f *FeatureAdapter) AddShapes(shape []*Shape) {
	f.Geometry = append(f.Geometry, shape.ToGeometry())
}
