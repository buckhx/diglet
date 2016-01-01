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

func (g *Geometry) ToLines() (lines []*Line) {
	cur := Point{X: 0, Y: 0}
	for _, cmd := range g.ToCommands() {
		switch cmd.cid {
		case MoveTo:
			x := cmd.params[0]
			y := cmd.params[1]
			cur = cur.Add(x, y)
			line := NewLine(cur)
			lines = append(lines, line)
			break
		case LineTo:
			x := cmd.params[0]
			y := cmd.params[1]
			cur = cur.Add(x, y)
			tail := lines[len(lines)-1]
			tail.Append(cur)
			break
		case ClosePath:
			tail := lines[len(lines)-1]
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

type Line struct {
	points []Point
}

func NewLine(points ...Point) *Line {
	return &Line{points}
}

func (l *Line) Append(point Point) {
	l.points = append(l.points, point)
}

func (l *Line) Equals(that *Line) bool {
	equal := len(l.points) == len(that.points)
	for i, point := range l.points {
		equal = equal && point == that.points[i]
	}
	return equal
}
