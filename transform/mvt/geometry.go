package mvt

import "fmt"

type state int

const (
	waitCint state = iota
	waitPint
)

const (
	MoveTo    = 1
	LineTo    = 2
	ClosePath = 7
)

type Geometry struct {
	geometry []uint
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
