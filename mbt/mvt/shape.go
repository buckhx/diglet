package mvt

import (
	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
)

type Point struct {
	X, Y int
}

func (p Point) Add(that Point) Point {
	x := p.X + that.X
	y := p.Y + that.Y
	return Point{X: x, Y: y}
}

func (p Point) Subtract(that Point) Point {
	x := p.X - that.X
	y := p.Y - that.Y
	return Point{X: x, Y: y}
}

func (p Point) Increment(x, y int) Point {
	return Point{X: p.X + x, Y: p.Y + y}
}

func (p Point) Decrement(x, y int) Point {
	return Point{X: p.X - x, Y: p.Y - y}
}

type CursorType string

const (
	AbsCur CursorType = "ABSOLUTE"
	RelCur CursorType = "RELATIVE"
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
		s.points[i] = p
	}
	return
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

func (s *Shape) GetPoints() []Point {
	return s.points
}

func (s *Shape) GetCursorType() CursorType {
	return s.curType
}

func (s *Shape) ToGeometrySlice() (geometry []uint32, err error) {
	chunks := make(chan []*command, 1000)
	go func() {
		defer close(chunks)
		head := 0
		commands := s.ToCommands()
		for cur := range commands {
			if commands[head].cid != commands[cur].cid {
				chunks <- commands[head:cur]
				head = cur
			}
		}
		chunks <- commands[head:len(commands)]
	}()
	for chunk := range chunks {
		geom, err := flushCommands(chunk)
		if err != nil {
			return nil, err
		}
		geometry = append(geometry, geom...)
	}
	return
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
