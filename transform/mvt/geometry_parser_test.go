package mvt

import "testing"

func TestReadPoints(t *testing.T) {
	tests := []geomTest{
		geomTest{[]uint{9, 50, 34}, []*command{
			&command{cid: moveTo, params: []int{25, 17}}},
		},
		geomTest{[]uint{17, 10, 14, 3, 9}, []*command{
			&command{cid: moveTo, params: []int{5, 7}},
			&command{cid: moveTo, params: []int{-2, -5}}},
		},
	}
	testGeometries(t, tests)
}

type geomTest struct {
	geometry []uint
	commands []*command
}

func testGeometries(t *testing.T, tests []geomTest) {
	for _, test := range tests {
		commands := []*command{}
		parser := &geomParser{state: waitCint}
		for _, gint := range test.geometry {
			cmd := parser.readGInt(gint)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
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
