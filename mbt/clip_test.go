package mbt

import (
	"testing"

	"github.com/buckhx/diglet/mbt/mvt"
)

func TestIntersection(t *testing.T) {
	tests := []struct {
		a, b []mvt.Point
		w    mvt.Point
	}{
		{
			a: []mvt.Point{{0, 0}, {0, 10}},
			b: []mvt.Point{{5, 5}, {-5, 5}},
			w: mvt.Point{0, 5},
		},
		{
			a: []mvt.Point{{0, 0}, {10, 10}},
			b: []mvt.Point{{10, 0}, {0, 10}},
			w: mvt.Point{5, 5},
		},
	}
	for _, test := range tests {
		p, _, _, err := intersection(test.a, test.b)
		if err != nil {
			t.Error(err)
		}
		if p != test.w {
			t.Errorf("Invalid Intersection wanted %v got %v", test.w, p)
		}
	}
}

func TestEdges(t *testing.T) {
	tests := []struct {
		s []mvt.Point
		e [][]mvt.Point
	}{
		{
			s: []mvt.Point{{0, 0}, {0, 10}, {10, 10}},
			e: [][]mvt.Point{{{0, 0}, {0, 10}}, {{0, 10}, {10, 10}}},
		},
		{
			s: []mvt.Point{{0, 0}, {0, 10}},
			e: [][]mvt.Point{{{0, 0}, {0, 10}}},
		},
		{
			s: []mvt.Point{{0, 0}},
			e: [][]mvt.Point{},
		},
	}
	for _, test := range tests {
		for i, e := range edges(test.s) {
			if e[0] != test.e[i][0] || e[1] != test.e[i][1] {
				t.Errorf("Invalid Edges wanted %v got %v", test.e, e)
				return
			}
		}
	}
}

func TestBoxContains(t *testing.T) {
	tests := []struct {
		p []mvt.Point //{nw, se, p}
		c bool
	}{
		{
			p: []mvt.Point{{0, 0}, {10, 10}, {5, 5}},
			c: true,
		},
		{
			p: []mvt.Point{{0, 0}, {10, 10}, {15, 15}},
			c: false,
		},
		{
			p: []mvt.Point{{0, 0}, {10, 10}, {0, 0}},
			c: true,
		},
		{
			p: []mvt.Point{{0, 0}, {10, 10}, {0, 5}},
			c: true,
		},
	}
	for _, test := range tests {
		b, err := bboxFromBounds(test.p[0], test.p[1])
		if err != nil {
			t.Error(err)
		}
		if test.c != b.contains(test.p[2]) {
			t.Errorf("Invalid BoxContains %v", test.p)
		}
	}
}

func TestBoxIntersect(t *testing.T) {
	tests := []struct {
		b, e []mvt.Point
		w    mvt.Point
	}{
		{
			b: []mvt.Point{{0, 0}, {10, 10}},
			e: []mvt.Point{{-5, 5}, {5, 5}},
			w: mvt.Point{0, 5},
		},
	}
	for _, test := range tests {
		b, err := bboxFromBounds(test.b[0], test.b[1])
		if err != nil {
			t.Error(err)
		}
		p, i := b.intersect(test.e)
		if i < 0 {
			t.Error("No box intersection")
		}
		if test.w != p {
			t.Errorf("Invalid BoxIntersect wanted %v got %v", test.w, p)
		}
	}
}

func TestClip(t *testing.T) {
	tests := []struct {
		b, s, w []mvt.Point
	}{
		{
			b: []mvt.Point{{0, -10}, {100, 100}},
			s: []mvt.Point{{-5, 15}, {5, 5}, {-5, -5}},
			w: []mvt.Point{{0, 10}, {5, 5}, {0, 0}, {0, 10}},
		},
		{
			b: []mvt.Point{{0, 0}, {10, 10}},
			s: []mvt.Point{{5, -5}, {5, 5}, {15, 5}},
			w: []mvt.Point{{5, 0}, {5, 5}, {10, 5}, {10, 0}, {5, 0}},
		},
		/*
			{
				b: []mvt.Point{{0, 0}, {10, 10}},
				s: []mvt.Point{{0, 0}, {5, 5}, {0, 10}},
				w: []mvt.Point{{0, 0}, {5, 5}, {0, 10}, {0, 0}},
			},
				{ // not intersecting
					b: []mvt.Point{{0, 0}, {10, 10}},
					s: []mvt.Point{{20, 20}, {50, 50}, {30, 30}},
					w: []mvt.Point{},
				},
				{ // shp inside of box
					b: []mvt.Point{{0, 0}, {10, 10}},
					s: []mvt.Point{{5, 5}, {6, 6}},
					w: []mvt.Point{{5, 5}, {6, 6}},
				},
				{ // shp contains of box
					b: []mvt.Point{{0, 0}, {10, 10}},
					s: []mvt.Point{{-100, -100}, {-100, 100}, {100, 100}, {100, -100}},
					w: []mvt.Point{{0, 0}, {0, 10}, {10, 10}, {10, 0}},
				},
		*/
	}
	for i, test := range tests {
		b, err := bboxFromBounds(test.b[0], test.b[1])
		if err != nil {
			t.Error(err)
		}
		c, err := b.clip(test.s)
		if err != nil {
			t.Error(err)
		}
		if len(c) != len(test.w) {
			t.Errorf("Invalid clip @ %d wanted %v got %v", i, test.w, c)
			return
		}
		for i, p := range c {
			if test.w[i] != p {
				t.Errorf("Invalid clip wanted %v got %v", test.w, c)
				return
			}
		}
	}
}
