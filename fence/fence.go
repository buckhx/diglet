package fence

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
)

const (
	RtreeFence      = "rtree"
	QuadRtreeFence  = "qrtree"
	QuadTreeFence   = "qtree"
	BruteForceFence = "brute"
)

type GeoFence interface {
	Add(f *geo.Feature)
	Get(c geo.Coordinate) []*geo.Feature
}

// Zoom only applies to q-based fences
func GetFence(fenceType string, zoom int) (fence GeoFence, err error) {
	switch fenceType {
	case RtreeFence:
		fence = NewRfence()
	case BruteForceFence:
		fence = NewBruteFence()
	case QuadTreeFence:
		fence = NewQfence(zoom)
	case QuadRtreeFence:
		fence = NewQrfence(zoom)
	default:
		err = util.Errorf("Bad fence type: %s", fenceType)
	}
	return
}
