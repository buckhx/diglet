package mvt

import "fmt"

type state int

const (
	waitCint state = iota
	waitPint
	cmdAvail
)

const (
	moveTo = 1
	lineTo = 2
	closeP = 7
)

type geomParser struct {
	state   state
	cmd     *command
	counter uint
}

func (p *geomParser) readGInt(gint uint) (cmd *command) {
	switch p.state {
	case waitCint:
		cid, cnt := readCmdInt(gint)
		p.counter = cnt
		p.cmd = &command{cid: cid, params: []int{}}
		p.state = waitPint
		break
	case waitPint:
		param := readPrmInt(gint)
		if p.cmd.needsParams() > 0 {
			p.cmd.appendParam(param)
		}
		break
	default:
		panic("Invalid state")

	}
	if p.cmd.needsParams() <= 0 {
		if p.counter > 0 {
			cmd = p.cmd
			p.cmd = &command{cid: p.cmd.cid, params: []int{}}
			p.counter--
		} else {
			cmd = p.cmd
			p.state = waitCint
		}
	}
	//fmt.Printf("%+v %+v\n", p, cmd)
	return
}

type command struct {
	cid    uint
	params []int
}

func (c *command) needsParams() uint {
	switch c.cid {
	case moveTo:
		return uint(2 - len(c.params))
	case lineTo:
		return uint(2 - len(c.params))
	default:
		return 0
	}
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
		1: "MoveTo",
		2: "LineTo",
		7: "CloseP",
	}
	return fmt.Sprintf("%s(%+v)", commands[c.cid], c.params)
}

func main() {
	//geometry := []uint{9, 50, 34}
	geometry := []uint{17, 10, 14, 3, 9}
	parser := &geomParser{state: waitCint}
	for _, gint := range geometry {
		cmd := parser.readGInt(gint)
		if cmd != nil {
			fmt.Printf("%s\n", cmd)
		}
	}
}

func readCmdInt(cmd uint) (cid, cnt uint) {
	cid = cmd & 0x7
	cnt = cmd >> 3
	return
}

func readPrmInt(pint uint) int {
	return int((pint >> 1) ^ (-(pint & 1)))
}
