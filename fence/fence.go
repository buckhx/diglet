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
	//Size() int
}

func GetFence(fenceType string) (fence GeoFence, err error) {
	switch fenceType {
	case RtreeFence:
		fence = NewRfence()
	case BruteForceFence:
		fence = NewBruteFence()
	case QuadTreeFence:
		fence = NewQfence(14) //using zoom==14 for neighborhood
		/*
			case QuadRtreeFence:
				fence = NewQrfence()
		*/
	default:
		err = util.Errorf("Bad fence type: %s", fenceType)
	}
	return
}
