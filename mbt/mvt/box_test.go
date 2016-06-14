package mvt

import "testing"

func TestBoxClip(t *testing.T) {
	tests := []struct {
		b    Box
		s, w []Point
	}{
		{
			b: Box{10, 10, 0, 0},
			s: []Point{{-5, 15}, {5, 5}, {-5, -5}, {-5, 15}},
			w: []Point{{0, 10}, {5, 5}, {0, 0}, {0, 10}},
		},
		{
			b: Box{10, 10, 0, 0},
			s: []Point{{-5, 5}, {15, 5}, {15, -5}, {-5, -5}, {-5, 5}},
			w: []Point{{0, 0}, {0, 5}, {10, 5}, {10, 0}, {0, 0}},
		},
	}
	for _, test := range tests {
		s := NewPolygon(append([]Point{}, test.s...)...)
		w := NewPolygon(test.w...)
		test.b.Clip(s)
		if !s.Equals(w) {
			t.Errorf("Invalid Box%v.Clip(%v) %v -> %v", test.b, test.s, test.w, s.points)
		}
	}
}

func TestBoxContains(t *testing.T) {
	tests := []struct {
		b Box
		p Point
		w bool
	}{
		{
			b: Box{10, 10, 0, 0},
			p: Point{5, 5},
			w: true,
		},
		{
			b: Box{10, 10, 0, 0},
			p: Point{0, 0},
			w: true,
		},
		{
			b: Box{10, 10, 0, 0},
			p: Point{-5, -5},
			w: false,
		},
	}
	for _, test := range tests {
		c := test.b.Contains(test.p)
		if c != test.w {
			t.Errorf("Invalid Box%v.Contains(%v) -> %v", test.b, test.p, c)
		}
	}
}

/*
func TestBoxCrosses(t *testing.T) {
	tests := []struct {
		b Box
		e edge
		w cross
	}{
		{
			b: Box{10, 10, 0, 0},
			e: edge{Point{5, 5}, Point{6, 6}},
			w: cross{},
		},
		{
			b: Box{10, 10, 0, 0},
			e: edge{Point{5, 5}, Point{5, 15}},
			w: cross{n: Point{5, 10}, tn: true},
		},
		{
			b: Box{10, 10, 0, 0},
			e: edge{Point{-15, 5}, Point{15, 5}},
			w: cross{e: Point{10, 5}, w: Point{0, 5}, te: true, tw: true}, //false, true, false, true},
		},
	}
	for _, test := range tests {
		c := test.b.crosses(test.e)
		if c != test.w {
			t.Errorf("Invalid Box%v.crosses(%v) %v -> %v", test.b, test.e, test.w, c)
		}
	}
}
*/
func TestEdgeLeft(t *testing.T) {
	tests := []struct {
		e edge
		p Point
		w bool
	}{
		{
			e: edge{Point{0, 0}, Point{0, 5}},
			p: Point{3, 3},
		},
		{
			e: edge{Point{0, 5}, Point{5, 5}},
			p: Point{3, 3},
		},
		{
			e: edge{Point{5, 5}, Point{5, 0}},
			p: Point{3, 3},
		},
		{
			e: edge{Point{5, 0}, Point{0, 0}},
			p: Point{3, 3},
		},
		{
			e: edge{Point{0, 0}, Point{0, 5}},
			p: Point{0, 3},
		},
		{
			e: edge{Point{0, 0}, Point{0, 5}},
			p: Point{-3, 300},
			w: true,
		},
	}
	for _, test := range tests {
		if test.e.left(test.p) != test.w {
			t.Errorf("Invalid edge%v.left(%v) %v -> %v", test.e, test.p, test.w, !test.w)
		}
	}
}

func TestEdgePCross(t *testing.T) {
	tests := []struct {
		e1, e2 edge
		w      Point
		ok     bool
	}{
		{
			e1: edge{Point{0, 0}, Point{10, 10}},
			e2: edge{Point{5, 0}, Point{5, 10}},
			w:  Point{5, 5},
			ok: true,
		},
		{
			e1: edge{Point{0, 0}, Point{0, 1}},
			e2: edge{Point{-5, 5}, Point{5, 5}},
			w:  Point{0, 5},
			ok: true,
		},
		{
			e1: edge{Point{0, 0}, Point{10, 10}},
			e2: edge{Point{0, 0}, Point{5, 0}},
		},
		{
			e1: edge{Point{0, 0}, Point{10, 10}},
			e2: edge{Point{0, 0}, Point{10, 10}},
		},
		{
			e1: edge{Point{10, 10}, Point{10, 0}},
			e2: edge{Point{0, 5}, Point{15, 5}},
			w:  Point{10, 5},
			ok: true,
		},
		/*
			{
				e1: edge{Point{0, 0}, Point{10, 10}},
				e2: edge{Point{0, 5}, Point{-5, -5}},
			},
		*/
	}
	for _, test := range tests {
		v, ok := test.e1.pcross(test.e2)
		if ok != test.ok || v != test.w {
			t.Errorf("Invalid edge%v.cross(%v) %v -> %v", test.e1, test.e2, test.w, v)
		}
	}
}
