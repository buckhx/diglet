package mvt

import (
	"fmt"

	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
)

type Shape struct {
	points   []Point
	curType  CursorType
	geomType vt.Tile_GeomType
}

func NewShape(geomType vt.Tile_GeomType, points ...Point) *Shape {
	return &Shape{points: points, curType: AbsCur}
}

func MakeShape(length int) *Shape {
	return &Shape{points: make([]Point, length), curType: AbsCur}
}

func NewPointShape(points ...Point) *Shape {
	return NewShape(vt.Tile_POINT, points...)
}

func NewLine(points ...Point) *Shape {
	return NewShape(vt.Tile_LINESTRING, points...)
}

func NewPolygon(points ...Point) *Shape {
	return NewShape(vt.Tile_POLYGON, points...)
}

func (s *Shape) Insert(i int, p Point) (err error) {
	if i >= len(s.points) || i < 0 {
		return util.Errorf("Insert index out of range %v @ %d", s.points, i)
	} else {
		a := s.points
		s.points = append(a[:i], append([]Point{p}, a[i:]...)...)
		//s.points[i] = p
	}
	return
}

func (s *Shape) Delete(i int) (p Point) {
	a := s.points
	p = a[i]
	s.points = append(a[:i], a[i+1:]...)
	return p
}

func (s *Shape) Append(point Point) {
	s.points = append(s.points, point)
}

func (s *Shape) Head() Point {
	return s.points[0]
}

func (s *Shape) Tail() Point {
	return s.points[len(s.points)-1]
}

func (s *Shape) Len() int {
	return len(s.points)
}

func (s *Shape) GetPoints() []Point {
	return s.points
}

func (s *Shape) GetCursorType() CursorType {
	return s.curType
}

func (s *Shape) ispoly() bool {
	fmt.Println(s.geomType)
	return s.geomType == vt.Tile_POLYGON
}

func (s *Shape) ToCommands() (cmds []*command) {
	switch s.geomType {
	case vt.Tile_POINT:
		move := newCmd(MoveTo, s.Head().X, s.Head().Y)
		cmds = []*command{move}
	case vt.Tile_LINESTRING:
		move := newCmd(MoveTo, s.Head().X, s.Head().Y)
		cmds = []*command{move}
		for _, p := range s.points[1:] {
			line := newCmd(LineTo, p.X, p.Y)
			cmds = append(cmds, line)
		}
	case vt.Tile_POLYGON:
		move := newCmd(MoveTo, s.Head().X, s.Head().Y)
		cmds = []*command{move}
		for _, p := range s.points[1:] {
			line := newCmd(LineTo, p.X, p.Y)
			cmds = append(cmds, line)
		}
		closep := newCmd(ClosePath)
		cmds = append(cmds, closep)
	default:
		util.Info("Unknown Geometry in ToCommands")
		cmds = []*command{}
	}
	return
}

/*
// Guess the shape by inspecting the points
func (s *Shape) SniffType() (gtype string) {
	if len(s.points) <= 0 {
		gtype = ShapeUNK
	} else if len(s.points) == 1 {
		gtype = ShapePNT
	} else if s.Head() == s.Tail() {
		gtype = ShapePLY
	} else if len(s.points) > 1 {
		gtype = ShapeLIN
	} else {
		gtype = ShapeUNK
	}
	return
}
*/

func (s *Shape) Equals(that *Shape) bool {
	equal := len(s.points) == len(that.points)
	for i, point := range s.points {
		equal = equal && point == that.points[i]
	}
	equal = equal && s.geomType == that.geomType
	return equal
}

func (s *Shape) edges() (edges []edge) {
	if s.Len() < 2 {
		return
	}
	edges = make([]edge, s.Len()-1)
	for i := 0; i < s.Len()-1; i++ {
		edges[i] = edge{s.points[i], s.points[i+1]}
	}
	return
}

func (s *Shape) edge(i int) edge {
	if i < 0 {
		i = s.Len() - i
	}
	return edge{s.points[i%s.Len()], s.points[(i+1)%(s.Len())]}
}

/*
func (s *Shape) infedges() (edges []edge) {
	edges = s.edges()
	for i, e := range edges {
		edges[i] = e.maxmult(1<<16, 1<<31-1)
	}
	return

}
*/
