package mvt

import (
	"reflect"
	"testing"
)

// These tests all come from the vector-tile-spec 2.0
// https://github.com/mapbox/vector-tile-spec/tree/master/2.0#435-example-geometry-encodings
func TestReadPoints(t *testing.T) {
	tests := []geomTest{
		geomTest{[]uint{9, 50, 34}, []*command{
			newCmd(MoveTo, 25, 17),
		}},
		geomTest{[]uint{17, 10, 14, 3, 9}, []*command{
			newCmd(MoveTo, 5, 7),
			newCmd(MoveTo, -2, -5),
		}},
	}
	testGeometries(t, tests)
}

func TestReadLine(t *testing.T) {
	tests := []geomTest{
		geomTest{[]uint{9, 4, 4, 18, 0, 16, 16, 0}, []*command{
			newCmd(MoveTo, 2, 2),
			newCmd(LineTo, 0, 8),
			newCmd(LineTo, 8, 0),
		}},
		geomTest{[]uint{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}, []*command{
			newCmd(MoveTo, 2, 2),
			newCmd(LineTo, 0, 8),
			newCmd(LineTo, 8, 0),
			newCmd(MoveTo, -9, -9),
			newCmd(LineTo, 2, 4),
		}},
	}
	testGeometries(t, tests)
}

func TestReadPolygons(t *testing.T) {
	tests := []geomTest{
		geomTest{[]uint{9, 6, 12, 18, 10, 12, 24, 44, 15}, []*command{
			newCmd(MoveTo, 3, 6),
			newCmd(LineTo, 5, 6),
			newCmd(LineTo, 12, 22),
			newCmd(ClosePath),
		}},
		//TODO test multi-polygon
		//https://github.com/mapbox/vector-tile-spec/tree/master/2.0#4356-example-multi-polygon
	}
	testGeometries(t, tests)
}

type geomTest struct {
	geometry []uint
	commands []*command
}

func testGeometries(t *testing.T, tests []geomTest) {
	for _, test := range tests {
		geometry := &Geometry{test.geometry}
		commands := geometry.ToCommands()
		if len(commands) != len(test.commands) {
			t.Errorf("Geometry parsing error %+v:\n\t%s ->\n\t%s",
				test.geometry, test.commands, commands)
		}
		for i, cmd := range test.commands {
			if !cmd.Equals(commands[i]) {
				t.Errorf("Geometry parsing error %+v at %d:\n\t%s ->\n\t%s",
					test.geometry, i, test.commands, commands)
			}
		}
	}
}

type shapeTest struct {
	geometry []uint
	shapes   []*Shape
}

func TestToPoints(t *testing.T) {
	tests := []shapeTest{
		shapeTest{[]uint{9, 50, 34}, []*Shape{
			NewShape(Point{25, 17}),
		}},
		shapeTest{[]uint{17, 10, 14, 3, 9}, []*Shape{
			NewShape(Point{5, 7}),
			NewShape(Point{3, 2}),
		}},
		shapeTest{[]uint{9, 4, 4, 18, 0, 16, 16, 0}, []*Shape{
			NewShape(Point{2, 2}, Point{2, 10}, Point{10, 10}),
		}},
		shapeTest{[]uint{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}, []*Shape{
			NewShape(Point{2, 2}, Point{2, 10}, Point{10, 10}),
			NewShape(Point{1, 1}, Point{3, 5}),
		}},
		shapeTest{[]uint{9, 6, 12, 18, 10, 12, 24, 44, 15}, []*Shape{
			NewShape(Point{3, 6}, Point{8, 12}, Point{20, 34}, Point{3, 6}),
		}},
	}
	for _, test := range tests {
		geom := &Geometry{test.geometry}
		shapes := geom.ToShapes()
		if len(shapes) != len(test.shapes) {
			t.Errorf("Geometry point translation error %+v:\n\t%s ->\n\t%s",
				test.geometry, test.shapes, shapes)
		}
		for i, shape := range test.shapes {
			if !shape.Equals(shapes[i]) {
				t.Errorf("Geometry point translation error %+v:\n\t%s ->\n\t%s",
					test.geometry, test.shapes, shapes)
			}
		}
	}
}

func TestGeometryRoundTrip(t *testing.T) {
	geometries := [][]uint32{
		[]uint32{9, 50, 34},
		[]uint32{17, 10, 14, 3, 9},
		[]uint32{9, 4, 4, 18, 0, 16, 16, 0},
		[]uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8},
		[]uint32{9, 6, 12, 18, 10, 12, 24, 44, 15},
	}
	for _, vtgeom := range geometries {
		geometry := GeometryFromVectorTile(vtgeom)
		shpgeom := []uint32{}
		for _, shape := range geometry.ToShapes() {
			slice, err := shape.ToGeometrySlice()
			if err != nil {
				t.Error(err)
			}
			shpgeom = append(shpgeom, slice...)
		}
		if !reflect.DeepEqual(shpgeom, vtgeom) {
			t.Errorf("Geometry did not round trip %v -> %v", vtgeom, shpgeom)
		}
	}
}

//TODO shape.GetType test
