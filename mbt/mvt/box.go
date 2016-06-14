package mvt

import "errors"

type Box struct {
	n, e, s, w int
}

func NewBox(sw, ne Point) (b Box, err error) {
	if sw.X >= ne.X || sw.Y >= ne.Y {
		err = errors.New("Box[0] >= Box[1]")
		return
	}
	b = Box{n: ne.Y, e: ne.X, s: sw.Y, w: sw.X}
	return
}

func (b Box) Contains(p Point) bool {
	n := p.Y <= b.n
	e := p.X <= b.e
	s := p.Y >= b.s
	w := p.X >= b.w
	return n && e && s && w
}

// Clip assumes that the union of b U s != nil (they overlap)
// if there is an intersection, c.points is backed by s.points
func (b Box) Clip(shp *Shape) {
	//TODO lot's of mallocs
	// move to shape
	out := shp.points
	ispoly := out[0] == out[len(out)-1]
	for _, e := range b.ToShape().edges() {
		in := append([]Point{}, out...)
		if len(in) < 1 {
			break
		}
		out = []Point{}
		s := in[len(in)-1]
		for _, v := range in {
			if e.inside(v) {
				if !e.inside(s) {
					a, _ := e.pcross(edge{s, v})
					out = append(out, a)
				}
				out = append(out, v)
			} else if e.inside(s) {
				a, _ := e.pcross(edge{s, v})
				out = append(out, a)
			}
			s = v
		}
	}
	if ispoly && len(out) > 2 {
		out = append(out, out[0])
	}
	shp.points = out
}

func (b Box) ToShape() *Shape {
	ne := Point{X: b.e, Y: b.n}
	se := Point{X: b.e, Y: b.s}
	sw := Point{X: b.w, Y: b.s}
	nw := Point{X: b.w, Y: b.n}
	return NewPolygon([]Point{sw, nw, ne, se, sw}...)
}
