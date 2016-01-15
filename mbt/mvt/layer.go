package mvt

import (
	"encoding/json"
	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
)

const VT_VERSION = 2

type Layer struct {
	vt_layer *vt.Tile_Layer
	keys     map[string]uint32
	values   map[*vt.Tile_Value]uint32
}

func newLayer(name string, extent uint) *Layer {
	var ver uint32 = VT_VERSION
	var ext uint32 = uint32(extent)
	vt_layer := &vt.Tile_Layer{
		Version:  &ver,
		Name:     &name,
		Extent:   &ext,
		Features: []*vt.Tile_Feature{},
		Keys:     []string{},
		Values:   []*vt.Tile_Value{},
	}
	return &Layer{
		vt_layer: vt_layer,
		keys:     make(map[string]uint32),
		values:   make(map[*vt.Tile_Value]uint32),
	}
}

func (l *Layer) AddFeature(feature *Feature) error {
	geom := feature.relativeGeometry().toVtGeometry()
	tags := l.tagProperties(feature.props)
	vt_feature := &vt.Tile_Feature{
		Id:       feature.id,
		Type:     &feature.gtype,
		Tags:     tags,
		Geometry: geom,
	}
	l.vt_layer.Features = append(l.vt_layer.Features, vt_feature)
	return nil
}

// Return a flat list of ordered pairs of indexes to the layers Keys and values
func (l *Layer) tagProperties(props map[string]interface{}) (tags []uint32) {
	tags = make([]uint32, 2*len(props))
	i := 0
	for key, val := range props {
		if _, ok := l.keys[key]; !ok {
			tag := uint32(len(l.vt_layer.Keys))
			l.vt_layer.Keys = append(l.vt_layer.Keys, key)
			l.keys[key] = tag
		}
		vt_val := getVtValue(val)
		if _, ok := l.values[vt_val]; !ok {
			tag := uint32(len(l.vt_layer.Values))
			l.vt_layer.Values = append(l.vt_layer.Values, vt_val)
			l.values[vt_val] = tag
		}
		tags[i] = l.keys[key]
		tags[i+1] = l.values[vt_val]
		i += 2
	}
	return
}

func getVtValue(val interface{}) (vt_val *vt.Tile_Value) {
	vt_val = &vt.Tile_Value{}
	switch v := val.(type) {
	case string:
		vt_val.StringValue = &v
	case float32:
		vt_val.FloatValue = &v
	case float64:
		vt_val.DoubleValue = &v
	case int:
		intv := int64(v)
		vt_val.IntValue = &intv
	case int32:
		intv := int64(v)
		vt_val.IntValue = &intv
	case int64:
		vt_val.IntValue = &v
	case uint:
		uintv := uint64(v)
		vt_val.UintValue = &uintv
	case uint32:
		uintv := uint64(v)
		vt_val.UintValue = &uintv
	case uint64:
		vt_val.UintValue = &v
	case bool:
		vt_val.BoolValue = &v
	default:
		//TODO, flatten maps
		err := util.Errorf("Bad interface{} for vt.Tile_Value  %v", val)
		util.Warn(err, "attempting to cast value to json string")
		if b, err := json.Marshal(v); err == nil {
			s := string(b)
			vt_val.StringValue = &s
		} else {
			util.Warn(err, "json cast failed, skipping...")
		}
	}
	return
}
