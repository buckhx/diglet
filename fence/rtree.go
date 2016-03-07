package fence

import (
	"github.com/buckhx/diglet/geo"
)

type Rfence struct {
	rtree *geo.Rtree
}

func NewRfence() *Rfence {
	return &Rfence{
		rtree: geo.NewRtree(),
	}
}

func (r *Rfence) Add(f *geo.Feature) {
	for _, shp := range f.Geometry {
		r.rtree.Insert(shp, f)
	}
}

func (r *Rfence) Get(c geo.Coordinate) []*geo.Feature {
	nodes := r.rtree.Contains(c)
	features := make([]*geo.Feature, len(nodes))
	for i, n := range nodes {
		features[i] = n.Feature()
	}
	return features
}
