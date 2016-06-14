package mvt

import (
	"strings"

	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
)

type Feature struct {
	//feature *vt.Tile_Feature
	gtype  vt.Tile_GeomType
	shapes []*Shape
	id     *uint64
	props  map[string]interface{}
}

func NewFeatureAdapter(id *uint64, geometry_type string, props map[string]interface{}) (feature *Feature) {
	feature = &Feature{id: id, shapes: []*Shape{}}
	feature.props = props
	switch strings.ToLower(geometry_type) {
	case "point", "multipoint":
		feature.gtype = vt.Tile_POINT
	case "linestring", "multilinestring":
		feature.gtype = vt.Tile_LINESTRING
	case "polygon", "multipolygon":
		feature.gtype = vt.Tile_POLYGON
	default:
		feature.gtype = vt.Tile_UNKNOWN
	}
	return
}

func (f *Feature) AddShape(shapes ...*Shape) {
	f.shapes = append(f.shapes, shapes...)
}

func (f *Feature) Valid() bool {
	return len(f.shapes) > 0
}

// MVT needs relative point instead of absolute
func (f *Feature) relativeGeometry() (geom *Geometry) {
	//TODO if RelCur -> skip translation
	geom = newGeometry(f.gtype, f.shapes...)
	cur := Point{X: 0, Y: 0}
	if f.id != nil {
		util.Debug("Feature Geometry id: %d", *f.id)
	} else {
		util.Debug("Feature Geometry <nil>")
	}
	for k, v := range f.props {
		util.Debug("\t%v: %v", k, v)
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
