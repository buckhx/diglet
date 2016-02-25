package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
	"sync"
)

func Excavate(q *Quarry, pbf, postcodes string) (err error) {
	util.Info("Excavating...")
	/*
		wg := &sync.WaitGroup{}
		//wg.Add(1)
		wg.Add(2)
		go survey(q, postcodes, wg)
		excavate(q, pbf, 8, wg)
		wg.Wait()
	*/
	subregions := loadSubregions(q)
	//util.Info("%s", subregions)
	//printGeojson(subregions)
	rtree := geo.NewRtree()
	for _, feature := range subregions {
		for _, shp := range feature.Geometry {
			rtree.Insert(shp, feature)
		}
	}
	//g := &sync.WaitGroup{}
	//for i := 0; i < 4; i++ {
	//g.Add(1)
	//go func() {
	//defer g.Done()
	counts := make(map[string]int, 88)
	for node := range q.Nodes() {
		c := geo.Coordinate{Lat: node.Lat, Lon: node.Lon}
		ins := rtree.Contains(c)
		if len(ins) == 0 {
			counts["No Relation"]++
			//util.Info("WARN: Node %d not contained", node.ID)
		} else {
			for _, rnode := range ins {
				feature := rnode.Feature()
				for _, shp := range feature.Geometry {
					if shp.Contains(c) {
						counts[feature.Tags("name")]++
					}
				}
			}
		}
	}
	//}()
	//}
	//g.Wait()
	total := 0
	for k, v := range counts {
		util.Info("%s: %d", k, v)
		total += v
	}
	util.Info("TOTAL: %d", total)
	return
}

func loadSubregions(q *Quarry) map[int64]*geo.Feature {
	subregions := make(map[int64]*geo.Feature) //, 88)
	for rel := range q.Relations() {
		/*
			if rel.ID != 1427734 {
				continue
			}
		*/
		if feature := relationFeature(q, rel); feature != nil {
			subregions[rel.ID] = feature
		}
	}
	return subregions
}

func survey(q *Quarry, postcodes string, wg *sync.WaitGroup) {
	defer wg.Done()
	q.Survey(postcodes)
}

func excavate(q *Quarry, pbf string, workers int, wg *sync.WaitGroup) {
	defer wg.Done()
	ex, err := osm.NewExcavator(pbf)
	util.Check(err)
	addrFilter := NewOsmFilter(1 << 27)
	ex.RelationCourier = func(feed <-chan *osm.Relation) {
		rels := make(chan QuarryRecord)
		go func() {
			defer close(rels)
			for rel := range feed {
				if rel.IsSubregionBoundary() {
					for _, m := range rel.Members {
						if m.Type == osm.WayType {
							addrFilter.AddInt64(m.ID)
						}
						rels <- rel
					}
				}
			}
		}()
		q.addRecords(RelationBucket, rels)
	}
	err = ex.Start(workers)
	util.Check(err)
	ex, err = osm.NewExcavator(pbf)
	util.Check(err)
	ex.WayCourier = func(feed <-chan *osm.Way) {
		ways := make(chan QuarryRecord)
		go func() {
			defer close(ways)
			for way := range feed {
				if way.IsAddressable() {
					addrFilter.AddInt64(way.ID)
					addrFilter.AddInt64(way.NodeIDs[0])
				}
				if addrFilter.HasInt64(way.ID) {
					for _, nid := range way.NodeIDs {
						addrFilter.AddInt64(nid)
					}
					ways <- way
				}
			}
		}()
		q.addRecords(WayBucket, ways)
	}
	ex.NodeCourier = func(feed <-chan *osm.Node) {
		for node := range feed {
			if node.IsAddressable() {
				addrFilter.AddInt64(node.ID)
			}
		}
	}
	err = ex.Start(workers)
	util.Check(err)
	ex, err = osm.NewExcavator(pbf)
	util.Check(err)
	ex.NodeCourier = func(feed <-chan *osm.Node) {
		nodes := make(chan QuarryRecord)
		go func() {
			defer close(nodes)
			for node := range feed {
				if addrFilter.HasInt64(node.ID) {
					nodes <- node
				}
			}
		}()
		q.addRecords(NodeBucket, nodes)
	}
	err = ex.Start(1)
	util.Check(err)
}

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
