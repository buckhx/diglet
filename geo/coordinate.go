package geo

import (
	"github.com/buckhx/diglet/util"
)

type Coordinate struct {
	Lat, Lon float64
}

func (c Coordinate) String() string {
	return util.Sprintf("[%.6f, %.6f]", c.Lat, c.Lon)
}
