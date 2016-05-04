package geo

import (
	"github.com/buckhx/tiles"
)

func QuadKey(c Coordinate, zoom int) tiles.Quadkey {
	return tiles.FromCoordinate(c.Lat, c.Lon, zoom).Quadkey()
}
