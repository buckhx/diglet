package mbt

import (
	"errors"
	"fmt"

	"github.com/buckhx/diglet/mbt/mvt"
)

type bbox []mvt.Point

func bboxFromBounds(sw, ne mvt.Point) (b bbox, err error) {
	if sw.X >= ne.X || sw.Y >= ne.Y {
		err = errors.New("Box[0] >= Box[1]")
		return
	}
	nw := sw.Increment(0, ne.Y-sw.Y)
	se := sw.Increment(ne.X-sw.X, 0)
	b = bbox{sw, nw, ne, se, sw}
	return
}

// Giler-Atherton
func (b bbox) clip(a []mvt.Point) (c []mvt.Point, err error) {
	// get intersections
	fmt.Println(b)
	type cross struct {
		p      mvt.Point
		ai, bi int
		am, bm int
		in     bool
	}
	xing := []cross{}
	for ai, ea := range edges(a) {
		for bi, eb := range edges(b) {
			if p, am, bm, err := intersection(ea, eb); err == nil {
				xing = append(xing, cross{p: p, ai: ai, bi: bi, am: am, bm: bm})
			}
		}
	}
	// TODO check for no intersections
	//insert intersections into slices and mark entries
	in := b.contains(a[0])
	for i, x := range xing {
		in = !in
		x.in = in
		x.ai = x.ai + i + 1
		x.bi = x.bi + i + 1
		xing[i] = x
		a = insert(a, x.p, x.ai, x.am)
		b = insert(b, x.p, x.bi, x.bm)
	}
	// traverse and build clipped shape
	for i, x := range xing {
		n := xing[0]
		if i < len(xing)-1 {
			n = xing[i+1]
		}
		if x.in { //entry take a
			d := (n.ai + 1) % len(a)
			c = append(c, a[x.ai:d]...)
		} else { //exit take box
			d := (n.bi + 1) % len(b)
			c = append(c, b[x.bi:d]...)
		}
	}
	c = append(c, c[0])
	return
}

func (b bbox) contains(p mvt.Point) bool {
	nw := b[0]
	se := b[2]
	nb := p.X >= nw.X && p.Y >= nw.Y
	sb := p.X <= se.X && p.Y <= se.Y
	return nb && sb
}

func (b bbox) edges() [][]mvt.Point {
	nw := b[0]
	ne := b[1]
	se := b[2]
	sw := b[3]
	n := []mvt.Point{nw, ne}
	e := []mvt.Point{ne, se}
	s := []mvt.Point{se, sw}
	w := []mvt.Point{sw, nw}
	return [][]mvt.Point{n, e, s, w}
}

// Returns the intersection point and edge index that it hit. If i < 0 there was no intersection.
func (b bbox) intersect(e []mvt.Point) (mvt.Point, int) {
	for i, be := range b.edges() {
		p, _, _, err := intersection(be, e)
		if err == nil {
			return p, i
		}
	}
	return mvt.Point{}, -1
}

// returns the intersections point of two segments, err if they don't intersect
// dist and b are ratio of how far along the line it is
func intersection(a, b []mvt.Point) (p mvt.Point, amag int, bmag int, err error) {
	if len(a) != 2 || len(b) != 2 {
		err = errors.New("Point slices are not line segments of len 2")
		return
	}
	ad := a[1].Subtract(a[0])
	bd := b[1].Subtract(b[0])
	axb := float64(ad.X*bd.Y - ad.Y*bd.X)
	if axb == 0 {
		err = errors.New("Line segments do not intersect")
		return
	}
	d := a[0].Subtract(b[0])
	ar := float64(bd.X*d.Y-bd.Y*d.X) / axb
	br := float64(ad.X*d.Y-ad.Y*d.X) / axb
	if ar < 0 || ar >= 1 || br < 0 || br >= 1 {
		err = errors.New("Line segments do not intersect, from ratios")
		return
	}
	ax := int(ar * float64(ad.X))
	ay := int(ar * float64(ad.Y))
	p = a[0].Increment(ax, ay)
	amag = dmag(a[0], p)
	bmag = dmag(b[0], p)
	return
}

// returns a slice of edges, which are len 2 []mvt.Point
func edges(a []mvt.Point) (e [][]mvt.Point) {
	s := len(a)
	if s < 2 {
		return
	}
	e = make([][]mvt.Point, s-1)
	for i := 1; i < s; i++ {
		e[i-1] = a[i-1 : i+1]
	}
	return
}

func insert(a []mvt.Point, v mvt.Point, i, mag int) []mvt.Point {
	for i > 1 { // will be before first
		if mag == dmag(a[i], v) {
			break
		}
		i--
	}
	return append(a[:i], append([]mvt.Point{v}, a[i:]...)...)
}

func reverse(a []mvt.Point) []mvt.Point {
	for i := len(a)/2 - 1; i >= 0; i-- {
		o := len(a) - 1 - i
		a[i], a[o] = a[o], a[i]
	}
	return a
}

// distance magnitude for comparison, not actual dist formula, but the relative magnitudes are correct for sorting
func dmag(a, b mvt.Point) int {
	d := b.Subtract(a)
	return d.X*d.X + d.Y*d.Y
}
