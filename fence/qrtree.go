package fence

import (
	"github.com/buckhx/diglet/geo"
)

type Qrfence struct {
	zoom   int
	qrtree map[string]*geo.Rtree
}

func NewQrfence(zoom int) *Qrfence {
	return &Qrfence{
		zoom:   zoom,
		qrtree: make(map[string]*geo.Rtree),
	}
}

func (q *Qrfence) Add(f *geo.Feature) {
	for _, shp := range f.Geometry {
		qkeys := shapeQkeys(shp, q.zoom)
		for _, key := range qkeys {
			if tree, ok := q.qrtree[key]; !ok {
				tree = geo.NewRtree()
				q.qrtree[key] = tree
				tree.Insert(shp, f)
			} else {
				tree.Insert(shp, f)
			}
		}
	}
}

func (q *Qrfence) Get(c geo.Coordinate) (matchs []*geo.Feature) {
	key := c.ToTile(q.zoom).QuadKey()
	if tree, ok := q.qrtree[key]; ok {
		nodes := tree.Contains(c)
		for _, n := range nodes {
			feature := n.Feature()
			if feature.Contains(c) {
				matchs = append(matchs, feature)
			}
		}
	}
	return
}
