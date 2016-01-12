package mvt

import (
	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
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

func (t *TileAdapter) GetTileBytes() ([]byte, error) {
	return vt.Encode(t.GetTile())
}

func (t *TileAdapter) GetTileGz() ([]byte, error) {
	return vt.EncodeGzipped(t.GetTile())
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

func (t *TileAdapter) AddFeature(layer string, adapter *FeatureAdapter) error {
	feature := &vt.Tile_Feature{
		Id:       adapter.id,
		Tags:     []uint32{},
		Type:     &adapter.gtype,
		Geometry: []uint32{},
	}
	geometry := adapter.RelativeGeometry()
	feature.Geometry = geometry.ToVtGeometry()
	l := t.layers[layer]
	l.Features = append(l.Features, feature)
	return nil
	//TODO add properties too, right now they'll be blank
}

type FeatureAdapter struct {
	//feature *vt.Tile_Feature
	gtype  vt.Tile_GeomType
	shapes []*Shape
	id     *uint64
}

func NewFeatureAdapter(id *uint64, geometry_type string) (adapter *FeatureAdapter) {
	adapter = &FeatureAdapter{id: id, shapes: []*Shape{}}
	switch strings.ToLower(geometry_type) {
	case "point", "multipoint":
		adapter.gtype = vt.Tile_POINT
	case "linestring", "multilinestring":
		adapter.gtype = vt.Tile_LINESTRING
	case "polygon", "multipolygon":
		adapter.gtype = vt.Tile_POLYGON
	default:
		adapter.gtype = vt.Tile_UNKNOWN
	}
	return
}

func (f *FeatureAdapter) AddShape(shapes ...*Shape) {
	f.shapes = append(f.shapes, shapes...)
}

// MVT needs relative point instead of absolute
func (f *FeatureAdapter) RelativeGeometry() (geom *Geometry) {
	//TODO if RelCur -> skip translation
	geom = NewGeometry(f.gtype, f.shapes...)
	cur := Point{X: 0, Y: 0}
	if f.id != nil {
		util.Debug("Feature Geometry id: %d", *f.id)
	} else {
		util.Debug("Feature Geometry <nil>")
	}
	for i, shape := range f.shapes {
		if f.gtype == vt.Tile_POLYGON {
			shape.points = shape.points[:len(shape.points)-1]
		}
		shape.geomType = f.gtype
		util.Debug("\tShape %d", i)
		for i, point := range shape.points {
			diff := point.Subtract(cur)
			shape.points[i] = diff
			cur = point
			util.Debug("\t\t%v", cur)
		}
		shape.curType = RelCur
	}
	return
}
