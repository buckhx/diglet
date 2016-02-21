package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
	"sync"
)

func Excavate(q *Quarry, pbf, postcodes string) (err error) {
	util.Info("Excavating...")
	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	//wg.Add(2)
	//go survey(q, postcodes, wg)
	//excavate(q, pbf, 8, wg)
	//wg.Wait()
	subregions := loadSubregions(q)
	printGeojson(subregions)
	//util.Info("%s", subregions)
	util.Info("%d", len(subregions))
	return
}

func loadSubregions(q *Quarry) map[int64]*geo.Feature {
	subregions := make(map[int64]*geo.Feature) //, 88)
	for rel := range q.Relations() {
		feature := geo.NewPolygonFeature()
		feature.Properties = make(map[string]interface{}, len(rel.Tags))
		for k, v := range rel.Tags {
			feature.Properties[k] = v
		}
		cur := geo.NewShape()
		shapes := make([]*geo.Shape, len(rel.Members))
		for i, m := range rel.Members {
			if m.Type != osm.WayType {
				// not going to deal with rel->rel & rel->node
				shapes[i] = geo.NewShape()
				continue
			}
			shapes[i] = wayShape(q, m.ID)
			/*
				if shp.Length() == 0 {
					util.Info("%d length 0, skipping", m.ID)
				} else if cur.Length() == 0 || cur.Tail() == shp.Head() {
					cur.Append(shp)
				} else if cur.Tail() == shp.Tail() {
					shp.Reverse()
					cur.Append(shp)
				} else {
					if cur.IsClosed() {
						switch m.Role {
						// outer rings -> clockwise
						// inner rings -> anti-clockwise
						case osm.RoleOuter:
							if !cur.IsClockwise() {
								cur.Reverse()
							}
						case osm.RoleInner:
							if cur.IsClockwise() {
								cur.Reverse()
							}
						}
					} else {
						//util.Info("Warning: Relation %d %q is not closed", rel.ID, rel.Tags["name"])
					}
					cur = shp
					feature.AddShape(cur)
				}
			*/
		}
		// Get the start and end shapes that have coordinates
		s := 0
		for i := 0; shapes[s].Length() == 0 && i < len(shapes); i++ {
			s = i
		}
		e := len(shapes) - 1
		for i := e; shapes[e].Length() == 0 && i > 0; i-- {
			e = i
		}
		// Flip the start if incorrectly wound
		if shapes[s].Tail() == shapes[e].Tail() || shapes[s].Tail() == shapes[e].Head() {
			shapes[s].Reverse()
		}
		cur = shapes[s]
		for i, shp := range shapes[s+1 : e+1] {
			c := s + i
			_ = c
			if shp.Length() == 0 {
				continue
			}
			if cur.Tail() == shp.Head() {
				cur.Add(shp.Coordinates[1:]...)
			} else if cur.Tail() == shp.Tail() {
				shp.Reverse()
				cur.Add(shp.Coordinates[1:]...)
			} else if shp.IsClosed() {
				feature.AddShape(cur)
				cur = shp
			} else {
				// there's a whole
				//util.Info("\tRelation %d has hole at %d", rel.ID, rel.Members[c].ID)
				cur.Add(shp.Coordinates[1:]...)
			}
		}
		feature.AddShape(cur)
		//todo close shape
		util.Info("Relation %d Geometry - %s", rel.ID, feature.Tags("wikipedia"))
		for i, shp := range feature.Geometry {
			if !shp.IsClosed() {
				util.Info("\tRelation %d not closed, explicitly closing", rel.ID)
				shp.Add(shp.Head())
			}
			util.Info("\t%d - %d - %t", i, shp.Length(), shp.IsClosed())

		}
		/*
			if len(feature.Geometry) == 0 {
				util.Info("Relation %d has nil geometry", rel.ID)
			} else {
				util.Info("%q IsClosed() -> %t", feature.Tags["name"], feature.Geometry[0].IsClosed())
			}
		*/
		/*
			if !feature.Geometry[len(feature.Geometry)-1].IsClosed() {
				util.Info("feature is not closed", feature.Tags["name"])
			}
		*/
		subregions[rel.ID] = feature
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
					//addrFilter.AddInt64(way.ID)
					//addrFilter.AddInt64(way.NodeIDs[0])
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
				//addrFilter.AddInt64(node.ID)
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

/*
	var wayc uint64 = 0
	var nidc uint64 = 0
	rels := cmap.New()
	ways := cmap.New()
	nods := cmap.New()
	addrFilter := NewOsmFilter(1 << 27)
	collectRelations := func(feed <-chan *Relation) {
		for rel := range feed {
			if rel.IsSubregionBoundary() {
				rels.Set(rel.Key(), rel)
			}
		}
	}
	collectWays := func(feed <-chan *Way) {
		for way := range feed {
			if way.IsSubregionBoundary() {
				ways.Set(way.Key(), way)
				for _, nid := range way.NodeIDs {
					nods.Set(strconv.FormatInt(nid, 10), nil)
				}
				nids := uint64(len(way.NodeIDs))
				atomic.AddUint64(&nidc, nids)
			}
			if way.IsAddressable() {
				atomic.AddUint64(&wayc, 1)
				addrFilter.AddInt64(way.ID)
				addrFilter.AddInt64(way.NodeIDs[0])
			}
		}
	}
*/

func wayShape(q *Quarry, wid int64) (shp *geo.Shape) {
	nodes, warns := q.WayNodes(wid)
	shp = geo.NewShape()
	for {
		select {
		case node, ok := <-nodes:
			if ok {
				c := geo.Coordinate{Lat: node.Lat, Lon: node.Lon}
				//util.Info("\t\t%d\t%s", node.ID, c)
				shp.Add(c)
			} else {
				return
			}
		case warn, ok := <-warns:
			if ok {
				_ = warn
				/*
					msg := "Incomplete relation %d %q, missing %d "
					if warn == m.ID {
						msg += "way"
					}
					util.Info(msg, rel.ID, rel.Tags["name"], warn)
				*/
			} else {
				//break nodeLoop
			}
		}
	}
	return
}
