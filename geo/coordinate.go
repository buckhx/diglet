package geo

import (
	"github.com/buckhx/diglet/util"
)

const (
	less     = -2
	lessOrEq = -1
	more     = 1
	moreOrEq = 2
)

type Coordinate struct {
	Lat, Lon float64
}

func (c Coordinate) X() float64 {
	return c.Lon
}

func (c Coordinate) Y() float64 {
	return c.Lat
}

func (c Coordinate) Difference(o Coordinate) (d Coordinate) {
	d.Lat = c.Lat - o.Lat
	d.Lon = c.Lon - o.Lon
	return
}

// Both coordinates are less than or equal
func (c Coordinate) strictCmp(o Coordinate) int {
	switch {
	case c.Lat < o.Lat && c.Lon < c.Lon:
		return less
	case c.Lat > o.Lat && c.Lon > c.Lat:
		return more
	case c.Lat <= o.Lat && c.Lon <= c.Lon:
		return lessOrEq
	case c.Lat >= o.Lat && c.Lon >= c.Lat:
		return moreOrEq
	default:
		return 0

	}
}

func (c Coordinate) String() string {
	return util.Sprintf("[%.6f, %.6f]", c.Lat, c.Lon)
}
