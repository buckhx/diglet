package geo

import (
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/rtreego"
)

var (
	RtreeMinChildren = 25
	RtreeMaxChildren = 50
	pointLen         = 0.00001 //~1m
)

type Rtree struct {
	rtree *rtreego.Rtree
}

func NewRtree() *Rtree {
	return &Rtree{
		rtree: rtreego.NewTree(RtreeMinChildren, RtreeMaxChildren),
	}
}

func (r *Rtree) Insert(s *Shape, data interface{}) {
	node := &Rnode{shape: s, data: data}
	r.rtree.Insert(node)
}

func (r *Rtree) Intersections(q *Shape) []*Rnode {
	query := rtreegoRect(q)
	return r.intersections(query)
}

func (r *Rtree) intersections(q *rtreego.Rect) []*Rnode {
	inodes := r.rtree.SearchIntersect(q)
	nodes := make([]*Rnode, len(inodes))
	for i, inode := range inodes {
		nodes[i] = inode.(*Rnode)
	}
	return nodes
}

func (r *Rtree) Contains(c Coordinate) []*Rnode {
	p := rtreego.Point{c.X(), c.Y()}
	rect := p.ToRect(pointLen)
	return r.intersections(rect)
}

type Rnode struct {
	shape *Shape
	data  interface{}
}

func (n *Rnode) Feature() *Feature {
	return n.Value().(*Feature)
}

func (n *Rnode) Value() interface{} {
	return n.data
}

//implements rtree.Spatial
func (n *Rnode) Bounds() *rtreego.Rect {
	return rtreegoRect(n.shape)
}

func rtreegoRect(s *Shape) *rtreego.Rect {
	bbox := s.BoundingBox()
	p := rtreego.Point{bbox.min.X(), bbox.min.Y()}
	d := bbox.max.Difference(bbox.min)
	rect, err := rtreego.NewRect(p, [2]float64{d.X(), d.Y()})
	util.Check(err)
	return &rect
}
