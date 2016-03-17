package geo

import (
	"github.com/buckhx/diglet/geo/tile_system"
	"github.com/buckhx/diglet/util"
	"math"
)

const (
	less     = -2
	lessOrEq = -1
	more     = 1
	moreOrEq = 2
	EarthRad = 6372800 //meters
	RadToDeg = 180 / math.Pi
	DegToRad = math.Pi / 180
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

//Distance between coordinates in meters
func (cd Coordinate) Distance(od Coordinate) float64 {
	c := cd.ToRad()
	o := od.ToRad()
	d := o.Difference(c)
	a := cos(c.Lat) * cos(o.Lat) * hvrsin(d.Lon)
	v := 2 * EarthRad * asin(sqrt(hvrsin(d.Lat)+a))
	return v
}

func (c Coordinate) ToRad() Coordinate {
	return Coordinate{Lat: c.Lat * DegToRad, Lon: c.Lon * DegToRad}
}

func (c Coordinate) ToTile(zoom int) (tile tile_system.Tile) {
	tile, _ = tile_system.CoordinateToTile(c.Lat, c.Lon, uint(zoom))
	return
}

// Both coordinates are less than or equal
func (c Coordinate) strictCmp(o Coordinate) int {
	switch {
	case c.Lat < o.Lat && c.Lon < o.Lon:
		return less
	case c.Lat > o.Lat && c.Lon > o.Lon:
		return more
	case c.Lat <= o.Lat && c.Lon <= o.Lon:
		return lessOrEq
	case c.Lat >= o.Lat && c.Lon >= o.Lon:
		return moreOrEq
	default:
		return 0
	}
}

func (c Coordinate) String() string {
	return util.Sprintf("[%.6f, %.6f]", c.Lat, c.Lon)
}

func sin(v float64) float64 {
	return math.Sin(v)
}

func cos(v float64) float64 {
	return math.Cos(v)
}

func asin(v float64) float64 {
	return math.Asin(v)
}

func sqrt(v float64) float64 {
	return math.Sqrt(v)
}

func hvrsin(v float64) float64 {
	return 0.5 * (1 - cos(v))
}
