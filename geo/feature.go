package geo

import (
	_ "github.com/buckhx/diglet/util"
)

const (
	PolygonFeature = "polygon"
	LineFeature    = "line"
	PointFeature   = "point"
)

type Feature struct {
	Geometry   []*Shape
	Type       string
	Properties map[string]interface{}
}

func NewFeature(geometryType string, geometry ...*Shape) *Feature {
	return &Feature{Geometry: geometry, Type: geometryType}
}

func NewPolygonFeature(geometry ...*Shape) *Feature {
	return NewFeature(PolygonFeature, geometry...)
}

func NewLineFeature(geometry ...*Shape) *Feature {
	return NewFeature(LineFeature, geometry...)
}

func NewPointFeature(geometry ...*Shape) *Feature {
	return NewFeature(PointFeature, geometry...)
}

func MakeFeature(length int) *Feature {
	return &Feature{Geometry: make([]*Shape, length)}
}

func (f *Feature) AddShape(s *Shape) {
	f.Geometry = append(f.Geometry, s)
}

func (f *Feature) Tags(key string) string {
	return f.Properties[key].(string)
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
