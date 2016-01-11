package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	ts "github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
)

type Coordinate struct {
	Lat, Lon float64
}

type Shape struct {
	Coordinates []Coordinate
}

func MakeShape(length int) *Shape {
	return &Shape{Coordinates: make([]Coordinate, length)}
}

func NewShape(coords ...Coordinate) *Shape {
	return &Shape{Coordinates: coords}
}

func (s *Shape) Append(c Coordinate) {
	s.Coordinates = append(s.Coordinates, c)
}

func (s *Shape) AddCoordinate(c Coordinate) {
	s.Coordinates = append(s.Coordinates, c)
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

// Unexported b/c the PointXY are still absolute, MVT needs relative
func (s *Shape) ToTileShape(tile ts.Tile) (shp *mvt.Shape) {
	shp = mvt.MakeShape(len(s.Coordinates))
	for i, c := range s.Coordinates {
		pixel := ts.ClippedCoords(c.Lat, c.Lon).ToPixel(tile.Z)
		x := int(pixel.X - tile.ToPixel().X)
		y := int(pixel.Y - tile.ToPixel().Y)
		point := mvt.Point{X: x, Y: y}
		//TODO: clipping
		shp.Insert(i, point)
	}
	return
}

type Feature struct {
	Id       *uint64
	Geometry []*Shape
	Type     string
	//Properties     *Metadata
}

func NewFeature(geometryType string, geometry ...*Shape) *Feature {
	return &Feature{Geometry: geometry, Type: geometryType}
}

func MakeFeature(length int) *Feature {
	return &Feature{Geometry: make([]*Shape, length)}
}

func (f *Feature) AddShape(s *Shape) {
	f.Geometry = append(f.Geometry, s)
}

func (f *Feature) SetF64Id(id float64) {
	var fid uint64 = uint64(id)
	f.Id = &fid
}

func (f *Feature) Center() (avg Coordinate) {
	div := 0.0
	avg = Coordinate{Lat: 0, Lon: 0}
	for _, shape := range f.Geometry {
		for _, c := range shape.Coordinates {
			avg.Lat += c.Lat
			avg.Lon += c.Lon
			div += 1
		}
	}
	avg.Lat /= div
	avg.Lon /= div
	return
}

func (f *Feature) ToTiledShapes(tile ts.Tile) (shps []*mvt.Shape) {
	shps = make([]*mvt.Shape, len(f.Geometry))
	for i, shape := range f.Geometry {
		shp := shape.ToTileShape(tile)
		shps[i] = shp
	}
	return
}

func (f *Feature) ToMvtAdapter(tile ts.Tile) (adapter *mvt.FeatureAdapter) {
	adapter = mvt.NewFeatureAdapter(f.Id, f.Type)
	adapter.AddShape(f.ToTiledShapes(tile)...)
	return
	//properties := featureValues(feature)
}
