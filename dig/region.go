package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
)

type rIndex struct {
	rtree *geo.Rtree
}

func (rdx *rIndex) getRelationKey(node *osm.Node) (key string) {
	c := geo.Coordinate{Lat: node.Lat, Lon: node.Lon}
	ins := rdx.rtree.Contains(c)
	l := len(ins)
	switch {
	case l == 1:
		key = ins[0].Feature().Tags("id")
	case l > 1:
		for _, rnode := range ins {
			feature := rnode.Feature()
			for _, shp := range feature.Geometry {
				if shp.Contains(c) {
					key = feature.Tags("id")
				}
			}
		}
	}
	return
}

//Load the regions from the db into a queryable index
func loadRIndex(q *Quarry) *rIndex {
	rdx := &rIndex{rtree: geo.NewRtree()}
	for rel := range q.Relations() {
		if feature := relationFeature(q, rel); feature != nil {
			for _, shp := range feature.Geometry {
				rdx.rtree.Insert(shp, feature)
			}
		}
	}
	return rdx
}

//Given an OSM relation, generate feature
func relationFeature(q *Quarry, rel *osm.Relation) (feature *geo.Feature) {
	feature = geo.NewPolygonFeature()
	feature.Properties = make(map[string]interface{}, len(rel.Tags))
	for k, v := range rel.Tags {
		feature.Properties[k] = v
	}
	//feature.Propertes["id"] = feature.ID
	ways := make(map[int64]*osm.Way, len(rel.Members))
	nodes := make(map[int64]geo.Coordinate, 3*len(rel.Members))
	for _, m := range rel.Members {
		if m.Type != osm.WayType {
			continue
		}
		way, nds := q.WayNodes(m.ID)
		if way == nil || len(nds) == 0 {
			util.Info("Missing member %d", m.ID)
			continue
		}
		ways[way.ID] = way
		for _, node := range nds {
			c := geo.Coordinate{Lat: node.Lat, Lon: node.Lon}
			nodes[node.ID] = c
		}
	}
	mems := rel.Members
	w0 := ways[mems[0].ID]
	w1 := ways[mems[1].ID]
	if w0 != nil && w1 != nil && w0.NodeIDs[0] == w1.NodeIDs[len(w1.NodeIDs)-1] {
		// hint at winding by reversing members
		for i, j := 0, len(mems)-1; i < j; i, j = i+1, j-1 {
			mems[i], mems[j] = mems[j], mems[i]
		}
	}
	popWay := func(ways map[int64]*osm.Way) (way *osm.Way) {
		for _, m := range rel.Members {
			way = ways[m.ID]
			if way != nil {
				delete(ways, way.ID)
				return
			}
		}
		return
	}
	wayShape := func(way *osm.Way) (shp *geo.Shape) {
		shp = geo.NewShape()
		for _, nid := range way.NodeIDs {
			shp.Add(nodes[nid])
		}
		return
	}
	nextWay := func(cur *osm.Way) *osm.Way {
		//head := way.NodeIDs[0]
		tail := cur.NodeIDs[len(cur.NodeIDs)-1]
		for _, way := range ways {
			if tail == way.NodeIDs[0] {
				delete(ways, way.ID)
				return way
			} else if tail == way.NodeIDs[len(way.NodeIDs)-1] {
				reverse(way.NodeIDs)
				delete(ways, way.ID)
				return way
			}
		}
		return nil
	}
	var shp *geo.Shape
	for len(ways) > 0 {
		way := popWay(ways)
		if shp == nil || shp.IsClosed() {
			//util.Info("----- New Shape -----")
			shp = geo.NewShape()
			feature.AddShape(shp)
		}
		for way != nil {
			//util.Info("%d:\t%v\t-> %v", way.ID, way.NodeIDs[0], way.NodeIDs[len(way.NodeIDs)-1])
			shp.Append(wayShape(way))
			way = nextWay(way)
		}
	}
	if !shp.IsClosed() {
		//return nil
		shp.Add(shp.Head())
	}
	return
}
