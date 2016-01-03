package mvt

import "fmt"

const (
	MoveTo    = 0x1
	LineTo    = 0x2
	ClosePath = 0x7
)

type state int

const (
	waitCint state = iota
	waitPint
)

type Geometry struct {
	geometry []uint
}

func (g *Geometry) ToShapes() (shapes []*Shape) {
	cur := Point{X: 0, Y: 0}
	for _, cmd := range g.ToCommands() {
		switch cmd.cid {
		case MoveTo:
			x := cmd.params[0]
			y := cmd.params[1]
			cur = cur.Add(x, y)
			shape := NewShape(cur)
			shapes = append(shapes, shape)
			break
		case LineTo:
			x := cmd.params[0]
			y := cmd.params[1]
			cur = cur.Add(x, y)
			tail := shapes[len(shapes)-1]
			tail.Append(cur)
			break
		case ClosePath:
			tail := shapes[len(shapes)-1]
			start := tail.points[0]
			tail.Append(start)
			break
		default:
			err := fmt.Errorf("Invalid CommandInteger %d", cmd.cid)
			panic(err)
		}
	}
	return
}

func (g *Geometry) ToCommands() (commands []*command) {
	var counter uint
	var cmd *command
	state := waitCint
	for _, gint := range g.geometry {
		switch state {
		case waitCint:
			cid, cnt := readCmdInt(gint)
			counter = cnt
			cmd = newCmd(cid)
			state = waitPint
			break
		case waitPint:
			param := readPrmInt(gint)
			if cmd.needsParams() > 0 {
				cmd.appendParam(param)
			} else {
				panic("waiting on param, but cmd doesn't need it")
			}
			break
		default:
			panic("Invalid state")
		}
		if cmd.needsParams() <= 0 {
			commands = append(commands, cmd)
			if counter > 0 {
				cmd = newCmd(cmd.cid)
				counter--
			} else {
				state = waitCint
			}
			if counter <= 0 {
				state = waitCint
			}
		}
	}
	return
}

func FromCommands(commands []*command) *Geometry {
	return nil
}

// Take a geometry slice from a pb-vector tile and convert it to []uint.
// This makes a copy which could be avoided if mvt.Geometry uses uint32 insted of uint
func GeometryFromVectorTile(vtgeom []uint32) *Geometry {
	geom := make([]uint, len(vtgeom))
	for i := range vtgeom {
		geom[i] = uint(vtgeom[i])
	}
	return &Geometry{geometry: geom}
}

type command struct {
	cid    uint
	params []int
}

func newCmd(cid uint, params ...int) *command {
	return &command{cid: cid, params: params}
}

func (c *command) needsParams() (count uint) {
	switch c.cid {
	case MoveTo:
		count = uint(2 - len(c.params))
	case LineTo:
		count = uint(2 - len(c.params))
	default:
		count = 0
	}
	return
}

func (c *command) appendParam(param int) {
	c.params = append(c.params, param)
}

func (c *command) Equals(that *command) bool {
	equal := c.cid == that.cid
	equal = equal && len(c.params) == len(that.params)
	for i, param := range c.params {
		equal = equal && param == that.params[i]
	}
	return equal
}

func (c *command) String() string {
	commands := map[uint]string{
		MoveTo:    "MoveTo",
		LineTo:    "LineTo",
		ClosePath: "ClosePath",
	}
	return fmt.Sprintf("%s(%+v)", commands[c.cid], c.params)
}

func readCmdInt(cmd uint) (cid, cnt uint) {
	cid = cmd & 0x7
	cnt = cmd >> 3
	return
}

func readPrmInt(pint uint) int {
	return int((pint >> 1) ^ (-(pint & 1)))
}

type Point struct {
	X, Y int
}

func (p Point) Add(x, y int) Point {
	return Point{X: p.X + x, Y: p.Y + y}
}

type Shape struct {
	points []Point
}

func NewShape(points ...Point) *Shape {
	return &Shape{points}
}

func (s *Shape) Append(point Point) {
	s.points = append(s.points, point)
}

func (s *Shape) GetPoints() []Point {
	return s.points
}

func (s *Shape) GetType() (gtype string) {
	if len(s.points) <= 0 {
		gtype = "UNKNOWN"
	} else if len(s.points) == 1 {
		gtype = "POINT"
	} else if s.points[0] == s.points[len(s.points)-1] {
		gtype = "POLYGON"
	} else if len(s.points) > 1 {
		gtype = "LINESTRING"
	} else {
		gtype = "UNKNOWN"
	}
	return
}

func (s *Shape) Equals(that *Shape) bool {
	equal := len(s.points) == len(that.points)
	for i, point := range s.points {
		equal = equal && point == that.points[i]
	}
	return equal
}
