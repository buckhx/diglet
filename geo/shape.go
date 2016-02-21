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
