package mbtiles

import (
	"strconv"
)

type Coordinate float64

func ParseCoordinate(coord string) (Coordinate, error) {
	c, err := strconv.ParseFloat(coord, 64)
	if err != nil {
		return Coordinate(0), err
	}
	return Coordinate(c), nil
}

type Bounds struct {
	left, bottom, right, top Coordinate
}
