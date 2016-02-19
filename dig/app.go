package dig

import (
	"github.com/buckhx/diglet/util"
	"sync"
)

func Excavate(q *Quarry, pbf, postcodes string) (err error) {
	util.Info("Excavating...")
	wg := &sync.WaitGroup{}
	//wg.Add(2)
	//go survey(q, postcodes, wg)
	excavate(q, pbf, 8, wg)
	//wg.Wait()
	return
}

func survey(q *Quarry, postcodes string, wg *sync.WaitGroup) {
	defer wg.Done()
	q.Survey(postcodes)
}

func excavate(q *Quarry, pbf string, workers int, wg *sync.WaitGroup) {
	defer wg.Done()
	ex, err := NewExcavator(pbf)
	util.Check(err)
	addrFilter := NewOsmFilter(1 << 27)
	ex.RelationCourier = func(feed <-chan *Relation) {
		rels := make(chan QuarryRecord)
		go func() {
			defer close(rels)
			for rel := range feed {
				if rel.IsSubregionBoundary() {
					rels <- rel
				}
			}
		}()
		q.addRecords(RelationBucket, rels)
	}
	ex.WayCourier = func(feed <-chan *Way) {
		ways := make(chan QuarryRecord)
		go func() {
			defer close(ways)
			for way := range feed {
				if way.IsAddressable() {
					addrFilter.AddInt64(way.ID)
					addrFilter.AddInt64(way.NodeIDs[0])
				}
				if way.IsSubregionBoundary() {
					for _, nid := range way.NodeIDs {
						addrFilter.AddInt64(nid)
					}
					ways <- way
				}
			}
		}()
		q.addRecords(WayBucket, ways)
	}
	ex.NodeCourier = func(feed <-chan *Node) {
		for node := range feed {
			if node.IsAddressable() {
				addrFilter.AddInt64(node.ID)
			}
		}
	}
	err = ex.Start(workers)
	util.Check(err)
	//util.Info("bloom cap: %d", addrFilter.filter.Capacity())
	ex, err = NewExcavator(pbf)
	util.Check(err)
	ex.NodeCourier = func(feed <-chan *Node) {
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
