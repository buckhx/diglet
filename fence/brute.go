package fence

import (
	"github.com/buckhx/diglet/geo"
)

type BruteFence struct {
	features []*geo.Feature
}

func NewBruteFence() *BruteFence {
	return &BruteFence{}
}

func (b *BruteFence) Add(f *geo.Feature) {
	b.features = append(b.features, f)
}

func (b *BruteFence) Get(c geo.Coordinate) []*geo.Feature {
	var ins []*geo.Feature
	for _, f := range b.features {
		for _, shp := range f.Geometry {
			if shp.Contains(c) {
				ins = append(ins, f)
			}
		}
	}
	return ins
}

func (b *BruteFence) Size() int {
	return len(b.features)
}
