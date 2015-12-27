package tile_system

import (
	"bytes"
	"strconv"
)

type Tile struct {
	X, Y, Z uint
}

func (t Tile) IntX() int {
	return int(t.X)
}

func (t Tile) IntY() int {
	return int(t.Y)
}

func (t Tile) ToPixel() Pixel {
	return Pixel{
		X: t.X * TileSize,
		Y: t.Y * TileSize,
		Z: t.Z,
	}
}

func (t Tile) ToPixelWithOffset(offset Pixel) Pixel {
	panic("Not Implemented")
	return Pixel{}
}

func (t Tile) QuadKey() string {
	var qk bytes.Buffer
	for i := t.Z; i > 0; i-- {
		quad := 0
		mask := 1 << (i - 1)
		if (t.IntX() & mask) != 0 {
			quad++
		}
		if (t.IntY() & mask) != 0 {
			quad += 2
		}
		digit := strconv.Itoa(quad)
		qk.WriteString(digit)
	}
	return qk.String()
}

func TileFromQuadKey(quadkey string) (tile Tile) {
	tile.Z = uint(len(quadkey))
	for i := tile.Z; i > 0; i-- {
		mask := uint(1 << (i - 1))
		cur := len(quadkey) - int(i)
		quad, err := strconv.Atoi(string(quadkey[cur]))
		check(err)
		switch uint(quad) {
		case 0:
			break
		case 1:
			tile.X |= mask
			break
		case 2:
			tile.Y |= mask
			break
		case 3:
			tile.X |= mask
			tile.Y |= mask
			break
		default:
			panic("Invalid quadkey " + quadkey)
		}
	}
	return
}
