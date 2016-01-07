package mvt

import (
	"fmt"
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
		x:      x,
		y:      y,
		z:      z,
		tile:   &vt.Tile{},
		layers: make(map[string]*vt.Tile_Layer),
	}
}

func (t *TileAdapter) GetTile() *vt.Tile {
	return t.tile
}

func (t *TileAdapter) AddLayer(name string, extent uint) {
	var ver uint32 = 2
	var ext uint32 = uint32(extent)
	layer := &vt.Tile_Layer{
		Version:  &ver,
		Name:     &name,
		Extent:   &ext,
		Features: []*vt.Tile_Feature{},
		Keys:     []string{},
		Values:   []*vt.Tile_Value{},
	}
	t.layers[name] = layer
	t.tile.Layers = append(t.tile.Layers, layer)
}

func (t *TileAdapter) AddFeature(layer string, feature *FeatureAdapter) {
	l := t.layers[layer]
	l.Features = append(l.Features, feature.feature)
	//TODO add properties too, right now they'll be blank
}

type FeatureAdapter struct {
	feature *vt.Tile_Feature
}

func NewFeatureAdapter(id *uint64, geometry_type string) *FeatureAdapter {
	var gtype vt.Tile_GeomType
	switch strings.ToLower(geometry_type) {
	case "point", "multipoint":
		gtype = vt.Tile_POINT
	case "linestring", "multilinestring":
		gtype = vt.Tile_LINESTRING
	case "polygon", "multipolygon":
		gtype = vt.Tile_POLYGON
	default:
		fmt.Println(gtype)
		gtype = vt.Tile_UNKNOWN
	}
	feature := &vt.Tile_Feature{
		Id:       id,
		Tags:     []uint32{},
		Type:     &gtype,
		Geometry: []uint32{},
	}
	return &FeatureAdapter{feature: feature}
}

func (f *FeatureAdapter) AddShapes(shapes []*Shape) {
	// Could save some space by flatenning MoveTo commands on contiguous ShapePNT
	// Adapt the code from shape.ToGeometry, but get a list of all []*Shape commands first
	for _, shape := range shapes {
		geom, _ := shape.ToGeometrySlice()
		f.feature.Geometry = append(f.feature.Geometry, geom...)
	}
}
