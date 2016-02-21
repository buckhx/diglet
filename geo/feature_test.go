package geo

import "testing"

func TestReverseShape(t *testing.T) {
	vals := []float64{5, 4, 3, 2, 1, 0}
	shape := NewShape()
	for i := range vals {
		f := float64(i)
		c := Coordinate{f, f}
		shape.Append(c)
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
