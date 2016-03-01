package geo

import (
	"github.com/buckhx/diglet/geo/tile_system"
)

func QuadKey(c Coordinate, zoom uint) string {
	coord := tile_system.ClippedCoords(c.Lat, c.Lon)
	pixel := coord.ToPixel(zoom)
	tile, _ := pixel.ToTile()
	return tile.QuadKey()
}
