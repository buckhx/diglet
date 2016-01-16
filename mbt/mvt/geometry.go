package mvt

import (
	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	"github.com/buckhx/diglet/util"
)

const (
	MoveTo    = 0x1
	LineTo    = 0x2
	ClosePath = 0x7
)

type Geometry struct {
	vt_type vt.Tile_GeomType
	shapes  []*Shape
	//vt_geom []uint
}

func newGeometry(vt_type vt.Tile_GeomType, shapes ...*Shape) *Geometry {
	return &Geometry{
		vt_type: vt_type,
		shapes:  shapes,
	}
}

func geometryFromVt(vt_type vt.Tile_GeomType, vtgeom []uint32) *Geometry {
	shapes, gtype, err := vtShapes(vtgeom)
	util.Check(err)
	if vt_type != gtype {
		util.Info("Assigned GeomType did not match sniffed %v -> %v", vt_type, gtype)
	}
	return newGeometry(vt_type, shapes...)
}

func (g *Geometry) toVtGeometry() (vtgeom []uint32) {
	cmds := make(chan *command, 100)
	go func() {
		defer close(cmds)
		for _, shape := range g.shapes {
			for _, cmd := range shape.ToCommands() {
				cmds <- cmd
			}
		}
	}()
	chunks := make(chan []*command, 100)
	go func() {
		defer close(chunks)
		chunk := []*command{}
		for cur := range cmds {
			if len(chunk) == 0 || chunk[0].cid == cur.cid {
				chunk = append(chunk, cur)
			} else {
				chunks <- chunk
				chunk = []*command{cur}
			}
		}
		chunks <- chunk
	}()
	for chunk := range chunks {
		geom, err := flushCommands(chunk)
		util.Check(err)
		vtgeom = append(vtgeom, geom...)
	}
	return
}

func vtShapes(geom []uint32) (shapes []*Shape, gtype vt.Tile_GeomType, err error) {
	cur := Point{X: 0, Y: 0}
	blocks := vtCommands(geom)
	shapes = make([]*Shape, len(blocks))
	for i, commands := range blocks {
		length := len(commands)
		if commands[len(commands)-1].cid == ClosePath {
			length--
		}
		shape := MakeShape(length)
		shapes[i] = shape
		for c, cmd := range commands {
			switch cmd.cid {
			case MoveTo:
				x := cmd.params[0]
				y := cmd.params[1]
				cur = cur.Increment(x, y)
				shape.points[c] = cur
				gtype = vt.Tile_POINT
				//shape := NewPointShape(cur)
				//shapes[i] = shape
				//i++
			case LineTo:
				x := cmd.params[0]
				y := cmd.params[1]
				cur = cur.Increment(x, y)
				shape.points[c] = cur
				gtype = vt.Tile_LINESTRING
				//shape := shapes[i-1]
				//shape.Append(cur)
			case ClosePath:
				gtype = vt.Tile_POLYGON
				//shape := shapes[len(shapes)-1]
				//start := shape.points[0]
				//shape.Append(start)
			default:
				err = util.Errorf("Invalid CommandInteger %d", cmd.cid)
				break
			}
		}
	}
	return
}

// Blocks of commands. Each block contains the commands for a shape
func vtCommands(vtgeom []uint32) (blocks [][]*command) {
	b := 0
	c := 0
	prmwait := false
	var cmdwait uint = 0
	for _, gint := range vtgeom {
		if prmwait {
			param := readPrmInt(gint)
			blocks[b][c].appendParam(param)
			prmwait = blocks[b][c].needsParams() > 0
			continue
		}
		if cmdwait > 0 {
			cmdwait--
			cid := blocks[b][c].cid
			if cid == MoveTo {
				blocks = append(blocks, []*command{})
				b = len(blocks) - 1
				c = -1
			}
			blocks[b] = append(blocks[b], newCmd(cid))
			c++
			prmwait = blocks[b][c].needsParams() > 0
			if prmwait {
				param := readPrmInt(gint)
				blocks[b][c].appendParam(param)
			}
		} else {
			cid, cnt := readCmdInt(gint)
			cmdwait = cnt - 1
			if cid == MoveTo {
				blocks = append(blocks, []*command{})
				b = len(blocks) - 1
				c = -1
			}
			blocks[b] = append(blocks[b], newCmd(cid))
			c++
		}
		prmwait = blocks[b][c].needsParams() > 0
	}
	return
}

type command struct {
	cid    uint
	params []int
}

func newCmd(cid uint, params ...int) *command {
	return &command{cid: cid, params: params}
}

func moveTo(x, y int) *command {
	return newCmd(MoveTo, x, y)
}

func lineTo(x, y int) *command {
	return newCmd(LineTo, x, y)
}

func closePath() *command {
	return newCmd(ClosePath)
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

func (c *command) paramIntegers() (pints []uint32) {
	for _, param := range c.params {
		pints = append(pints, writePrmInt(param))
	}
	return
}

func (c *command) Equal(that *command) bool {
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
	return util.Sprintf("%s(%+v)", commands[c.cid], c.params)
}

func readCmdInt(cmd uint32) (cid, cnt uint) {
	c := uint(cmd)
	cid = c & 0x7
	cnt = c >> 3
	return
}

func readPrmInt(pint uint32) int {
	p := uint(pint)
	return int((p >> 1) ^ (-(p & 1)))
}

func writeCmdInt(cid, cnt uint) uint32 {
	return uint32((cid & 0x7) | (cnt << 3))
}

func writePrmInt(param int) uint32 {
	return uint32((param << 1) ^ (param >> 31))
}

func flushCommands(chunk []*command) (geom []uint32, err error) {
	if len(chunk) < 1 {
		err = util.Errorf("Flushing Zero-Length command chunk")
	} else {
		cid := chunk[0].cid
		cnt := uint(len(chunk))
		cint := writeCmdInt(cid, cnt)
		geom = append(geom, cint)
		for _, cmd := range chunk {
			if cmd.cid != cid {
				msg := "Non contiguous CommandInteger in command chunk: %v"
				err = util.Errorf(msg, chunk)
				return
			}
			for _, param := range cmd.params {
				pint := writePrmInt(param)
				geom = append(geom, pint)
			}
		}
	}
	return
}
