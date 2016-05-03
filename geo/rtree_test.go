package geo_test

import (
	"github.com/buckhx/diglet/geo"
	"testing"
)

func TestRtreeIntersection(t *testing.T) {
	points := [][]geo.Coordinate{
		{{0, 0}, {1, 1}},
		{{-1, -1}, {0, 0}},
		{{-1, -1}, {1, 1}},
		{{5, 5}, {6, 6}},
	}
	rtree := geo.NewRtree()
	for i, ps := range points {
		shp := geo.NewShape(ps...)
		rtree.Insert(shp, i)
	}
	q := cd(0.5, 0.5)
	if len(rtree.Contains(q)) != 2 {
		t.Errorf("Failed rtree containment: %d wanted 2", len(rtree.Contains(q)))

	}

}

func cd(x, y float64) geo.Coordinate {
	return geo.Coordinate{x, y}
}

/*
func TestClockwise(t *testing.T) {
	antiwise := []Coordinate{
		Coordinate{39.7435437641, -105.003612041},
		Coordinate{39.7427848013, -105.003011227},
		Coordinate{39.7431642838, -105.002217293},
		Coordinate{39.7439067434, -105.002839565},
		Coordinate{39.7435437641, -105.003612041},
	}
	shape := NewShape(antiwise...)
	if shape.IsClockwise() {
		t.Errorf("Shape is clockwise %v", shape)
	}
	shape.Reverse()
	if !shape.IsClockwise() {
		t.Errorf("Shape is not clockwise %v", shape)
	}
}
*/
