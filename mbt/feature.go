package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	ts "github.com/buckhx/diglet/mbt/tile_system"
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

func (s *Shape) ToMvtShape(zoom uint) (shp *mvt.Shape) {
	shp = mvt.MakeShape(len(s.Coordinates))
	for i, c := range s.Coordinates {
		pixel := ts.ClippedCoords(c.Lat, c.Lon).ToPixel(zoom)
		tile, _ := pixel.ToTile()
		origin := tile.ToPixel()
		x := int(pixel.X - origin.X)
		y := int(pixel.Y - origin.Y)
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

func (f *Feature) ToMvtShapes(zoom uint) (shps []*mvt.Shape) {
	shps = make([]*mvt.Shape, len(f.Geometry))
	for i, shape := range f.Geometry {
		shp := shape.ToMvtShape(zoom)
		shps[i] = shp
	}
	return shps
}

func (f *Feature) ToMvtAdapter(zoom uint) (adapter *mvt.FeatureAdapter) {
	adapter = mvt.NewFeatureAdapter(f.Id, f.Type)
	adapter.AddShapes(f.ToMvtShapes(zoom))
	return
	//properties := featureValues(feature)
}
