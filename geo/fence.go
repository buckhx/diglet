package geo

type Fence interface {
	Add(f *Feature)
	GetFeatures(f *Feature) []*Feature
	Size() int
}

type Rfence struct {
	rtree *Rtree
}

func NewRfence() *Rfence {
	return &Rfence{
		rtree: NewRtree(),
	}
}

func (r *Rfence) Add(f *Feature) {
	for _, shp := range f.Geometry {
		r.rtree.Insert(shp, f)
	}
}

func (r *Rfence) Get(c Coordinate) []*Feature {
	nodes := r.rtree.Contains(c)
	features := make([]*Feature, len(nodes))
	for i, n := range nodes {
		features[i] = n.data.(*Feature)
	}
	return features
}

func (r *Rfence) Size() int {
	return r.rtree.rtree.Size()
}

type BruteFence struct {
	features []*Feature
}

func NewBruteFence() *BruteFence {
	return &BruteFence{}
}

func (b *BruteFence) Add(f *Feature) {
	b.features = append(b.features, f)
}

func (b *BruteFence) Get(c Coordinate) []*Feature {
	var ins []*Feature
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
