package geo

import (
	"github.com/buckhx/diglet/util"
)

type Shape struct {
	Coordinates []Coordinate
}

func MakeShape(length int) *Shape {
	return &Shape{Coordinates: make([]Coordinate, length)}
}

func NewShape(coords ...Coordinate) *Shape {
	return &Shape{Coordinates: coords}
}

func (s *Shape) Append(o *Shape) {
	s.Coordinates = append(s.Coordinates, o.Coordinates...)
}

func (s *Shape) Add(c ...Coordinate) {
	s.Coordinates = append(s.Coordinates, c...)
}

func (s *Shape) BoundingBox() Box {
	h := s.Head()
	min, max := h, h
	for _, c := range s.Coordinates {
		if c.Lat < min.Lat {
			min.Lat = c.Lat
		}
		if c.Lat > max.Lat {
			max.Lat = c.Lat
		}
		if c.Lon < min.Lon {
			min.Lon = c.Lon
		}
		if c.Lon > max.Lon {
			max.Lon = c.Lon
		}
	}
	box, _ := NewBox(min, max)
	return box
}

func (s *Shape) Head() Coordinate {
	return s.Coordinates[0]
}

func (s *Shape) Tail() Coordinate {
	return s.Coordinates[s.Length()-1]
}

// Only shapes that contain an area can be closed (>3 Coordinates)
func (s *Shape) IsClosed() bool {
	if s.Length() < 4 {
		return false
	}
	return s.Coordinates[0] == s.Coordinates[s.Length()-1]
}

func (s *Shape) Reverse() {
	last := len(s.Coordinates) - 1
	for i, c := range s.Coordinates {
		s.Coordinates[i] = s.Coordinates[last-i]
		s.Coordinates[last-i] = c
		if i >= last/2 {
			break
		}
	}
}

// Number of coordinates
func (s *Shape) Length() int {
	return len(s.Coordinates)
}

func (s *Shape) IsClockwise() bool {
	sum := 0.0
	for i, c := range s.Coordinates[:len(s.Coordinates)-1] {
		n := s.Coordinates[i+1]
		sum += (n.Lon - c.Lon) * (n.Lat + c.Lat)
	}
	if sum == 0 {
		util.Info("Shape edge sum == 0, defaulting to clockwise == true")
	}
	return sum > 0
}

type Box struct {
	min, max Coordinate
}

func NewBox(min, max Coordinate) (box Box, err error) {
	if min.Lat > max.Lat || min.Lon > max.Lon {
		err = util.Errorf("Min %v > Max %v", min, max)
	} else {
		box = Box{min: min, max: max}
	}
	return
}

func (b Box) Contains(coords ...Coordinate) (in bool) {
	for _, c := range coords {
		in = in || (b.min.strictCmp(c) < 0 && b.max.strictCmp(c) > 0)
		if in {
			return
		}
	}
	return
}

//TODO implement this
/*
func (b *Box) Intersect(o *Box) *Box {
}
*/
