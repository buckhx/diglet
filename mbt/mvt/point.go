package mvt

import "math"

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

type edge struct {
	v1, v2 Point
}

func (e edge) diff() Point {
	return e.v2.Subtract(e.v1)
}

func (e edge) boxed(p Point) bool {
	var n, r, s, w int
	if e.v1.Y > e.v2.Y {
		n = e.v1.Y
		s = e.v2.Y
	} else {
		s = e.v1.Y
		n = e.v2.Y
	}
	if e.v1.X > e.v2.X {
		r = e.v1.X
		w = e.v2.X
	} else {
		w = e.v1.X
		r = e.v2.X
	}
	return n >= p.Y && s <= p.Y && r >= p.X && w <= p.X
}

func (e edge) pcross(o edge) (v Point, ok bool) {
	m := e.slope()
	a := o.slope()
	if m == a {
		return
	}
	ey, ex := float64(e.v1.Y), float64(e.v1.X)
	oy, ox := float64(o.v1.Y), float64(o.v1.X)
	b := ey - m*ex
	d := oy - a*ox
	x := (b - d) / (a - m)
	y := m*x + b
	if m > math.MaxFloat32 || m < -1*math.MaxFloat32 {
		// handle vertical
		y = a*x + d
	}
	v = Point{X: int(x), Y: int(y)}
	eq := v == e.v1 || v == e.v2 || v == o.v1 || v == o.v2
	ok = o.boxed(v) && !eq
	return
}

func (e edge) left(p Point) bool {
	d := e.diff()
	return (d.X)*(p.Y-e.v1.Y) > (d.Y)*(p.X-e.v1.X)
}

func (e edge) inside(p Point) bool {
	return !e.left(p)
}

func (e edge) slope() float64 {
	dy := float64(e.v2.Y - e.v1.Y)
	dx := float64(e.v2.X - e.v1.X)
	if dx == 0 {
		dx = 1 / math.MaxFloat32
	}
	return dy / dx
}

func (e edge) cross(o edge) (v Point, ok bool) {
	ad := e.diff()
	bd := o.diff()
	axb := float64(ad.X*bd.Y - ad.Y*bd.X)
	if axb == 0 {
		return
	}
	d := e.v1.Subtract(o.v1)
	ar := float64(bd.X*d.Y-bd.Y*d.X) / axb
	br := float64(ad.X*d.Y-ad.Y*d.X) / axb
	if ar < 0 || ar >= 1 || br < 0 || br >= 1 {
		return
	}
	ax := int(ar * float64(ad.X))
	ay := int(ar * float64(ad.Y))
	v = e.v1.Increment(ax, ay)
	return v, true
}

/*

// find the point where this edge interests with the x plane if it were a line
func (e edge) xplane(x int) Point {
	m := e.slope()
	b := e.v1.Y - m*e.v1.X
	y := m*x + b
	return Point{X: x, Y: y}
}

// find the point where this edge interests with the x plane if it were a line
func (e edge) yplane(y int) Point {
	o := edge{Point{e.v1.Y, e.v1.X}, Point{e.v2.Y, e.v2.X}}
	p := o.xplane(y)
	return Point{p.Y, p.X}
}

func (e edge) xcross(x int) (p Point, ok bool) {
	if (e.v1.X <= x && e.v2.X >= x) || (e.v2.X <= x && e.v1.X >= x) {
		return
	}
	m := e.slope()
	b := e.v1.Y - m*e.v1.X
	y := m*x + b
	return Point{X: x, Y: y}, true
}

func (e edge) ycross(y int) (p Point, ok bool) {
	if (e.v1.Y <= y && e.v2.Y >= y) || (e.v2.Y <= y && e.v1.Y >= y) {
		return
	}
	o := edge{Point{e.v1.Y, e.v1.X}, Point{e.v2.Y, e.v2.X}}
	p, ok = o.xcross(y)
	return Point{p.Y, p.X}, ok
}
*/
