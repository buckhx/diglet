package mvt

import (
	//vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	//"reflect"
	"testing"
)

// These tests all come from the vector-tile-spec 2.0
// https://github.com/mapbox/vector-tile-spec/tree/master/2.0#435-example-geometry-encodings
func TestReadPoints(t *testing.T) {
	tests := []geomTest{
		{[]uint32{9, 50, 34}, []*command{
			newCmd(MoveTo, 25, 17),
		}},
		{[]uint32{17, 10, 14, 3, 9}, []*command{
			newCmd(MoveTo, 5, 7),
			newCmd(MoveTo, -2, -5),
		}},
	}
	testGeometries(t, tests)
}

func TestReadLine(t *testing.T) {
	tests := []geomTest{
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0}, []*command{
			newCmd(MoveTo, 2, 2),
			newCmd(LineTo, 0, 8),
			newCmd(LineTo, 8, 0),
		}},
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}, []*command{
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
		{[]uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}, []*command{
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
	geometry []uint32
	commands []*command
}

func testGeometries(t *testing.T, tests []geomTest) {
	for _, test := range tests {
		//geometry := GeometryFromVt(vt.Tile_UNKNOWN, test.geometry)
		blocks := vtCommands(test.geometry)
		i := 0
		for _, commands := range blocks {
			for _, cmd := range commands {
				if !cmd.Equal(test.commands[i]) {
					msg := "Geometry parsing error %+v at %d:\n\t%s ->\n\t%s"
					t.Errorf(msg, test.geometry, i, test.commands, blocks)
					//t.Errorf(msg, test.geometry, i, test.commands[i], cmd)
				}
				i++
			}
		}
		if i != len(test.commands) {
			msg := "GeometryLen parsing error %+v at %d:\n\t%s ->\n\t%s"
			t.Errorf(msg, test.geometry, i, test.commands, blocks)
		}
	}
}

type shapeTest struct {
	geometry []uint32
	shapes   []*Shape
}

func TestToShapes(t *testing.T) {
	tests := []shapeTest{
		{[]uint32{9, 50, 34}, []*Shape{
			NewPointShape(Point{25, 17}),
		}},
		{[]uint32{17, 10, 14, 3, 9}, []*Shape{
			NewPointShape(Point{5, 7}),
			NewPointShape(Point{3, 2}),
		}},
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0}, []*Shape{
			NewLine(Point{2, 2}, Point{2, 10}, Point{10, 10}),
		}},
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}, []*Shape{
			NewLine(Point{2, 2}, Point{2, 10}, Point{10, 10}),
			NewLine(Point{1, 1}, Point{3, 5}),
		}},
		{[]uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}, []*Shape{
			NewPolygon(Point{3, 6}, Point{8, 12}, Point{20, 34}),
		}},
	}
	for _, test := range tests {
		shapes, _, err := vtShapes(test.geometry)
		if err != nil {
			t.Error(err)
		}
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

/*
func TestDerp(t *testing.T) {
	mpoly := []*command{
		moveTo(0, 0),
		lineTo(10, 0),
		lineTo(0, 10),
		lineTo(-10, 0),
		closePath(),
		moveTo(11, 1),
		lineTo(9, 0),
		lineTo(0, 9),
		lineTo(9, 0),
		closePath(),
		moveTo(2, -7),
		lineTo(0, 4),
		lineTo(4, 0),
		lineTo(0, 4),
		closePath(),
	}

}
*/

/*
func TestGeometryRoundTrip(t *testing.T) {
	geomTests := []struct {
		geom  []uint32
		gtype vt.Tile_GeomType
	}{
		{[]uint32{9, 50, 34}, vt.Tile_POINT},
		{[]uint32{17, 10, 14, 3, 9}, vt.Tile_POINT},
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0}, vt.Tile_LINESTRING},
		{[]uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8}, vt.Tile_LINESTRING},
		{[]uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}, vt.Tile_POLYGON},
	}
	for _, test := range geomTests {
		geometry := GeometryFromVt(test.gtype, test.geom)
		shpgeom := []uint32{}
		for _, shape := range geometry.shapes {
			slice, err := shape.ToGeometrySlice()
			if err != nil {
				t.Error(err)
			}
			shpgeom = append(shpgeom, slice...)
		}
		if !reflect.DeepEqual(shpgeom, test.geom) {
			t.Errorf("Geometry did not round trip %v -> %v", test.geom, shpgeom)
		}
	}
}
*/
