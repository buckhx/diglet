package mvt

import "testing"

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
