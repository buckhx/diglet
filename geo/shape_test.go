package geo

import "testing"

func TestReverseShape(t *testing.T) {
	vals := []float64{5, 4, 3, 2, 1, 0}
	shape := NewShape()
	for i := range vals {
		f := float64(i)
		c := Coordinate{f, f}
		shape.Add(c)
	}
	shape.Reverse()
	for i, c := range shape.Coordinates {
		if vals[i] != c.Lat {
			t.Errorf("Shape was not reversed %v", shape)
		}
	}
}

func TestClockwise(t *testing.T) {
	antiwise := []Coordinate{
		{39.7435437641, -105.003612041},
		{39.7427848013, -105.003011227},
		{39.7431642838, -105.002217293},
		{39.7439067434, -105.002839565},
		{39.7435437641, -105.003612041},
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

func TestBBox(t *testing.T) {
	boxes := []Box{
		{min: cd(0, 0), max: cd(10, 10)},
	}
	points := [][]Coordinate{
		{
			cd(0, 5), cd(5, 0), cd(5, 10), cd(10, 5),
		},
	}
	for i, coords := range points {
		shp := NewShape(coords...)
		if shp.BoundingBox() != boxes[i] {
			t.Errorf("Write BBox test %v", shp.BoundingBox())

		}
	}
}

func TestContains(t *testing.T) {
	points := []Coordinate{cd(5, 5)}
	shapes := [][]Coordinate{
		{
			cd(0, 5), cd(5, 0), cd(5, 10), cd(10, 5),
		},
	}
	for i, coords := range shapes {
		shp := NewShape(coords...)
		if !shp.Contains(points[i]) {
			t.Errorf("Shape !contains %v %v", shp, points[i])
		}
	}
}

func cd(x, y float64) Coordinate {
	return Coordinate{x, y}
}

func TestDistance(t *testing.T) {
	x := Coordinate{Lat: 36.12, Lon: -86.67}
	y := Coordinate{Lat: 33.94, Lon: -118.40}
	d := int(x.Distance(y))
	if d != 2887259 {
		t.Errorf("Bad distance %s.Distance(%s) != %f", x, y, d)
	}
}
