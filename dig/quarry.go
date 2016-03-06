package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
	"sync"
)

type Quarry struct {
	db  *Qdb
	rdx *rIndex
}

type Match struct {
	Query  Address
	Result Address
	Meta   map[string]string
}

func (m Match) String() string {
	qry := m.Query.String()
	res := m.Result.String()
	lat := m.Result.Location.Lat
	lon := m.Result.Location.Lon
	qkey := geo.QuadKey(m.Result.Location, 23)
	return util.Sprintf("%q,%q,\"%f\",\"%f\",%q,\"%f\"", qry, res, lat, lon, qkey, m.Edist())
}

// Edit distance between query and result
func (m Match) Edist() float64 {
	return m.Query.dist(m.Result)
}

func OpenQuarry(path string) (q *Quarry, err error) {
	db, err := OpenQdb(path)
	if err != nil {
		return
	}
	rdx := loadRIndex(db)
	q = &Quarry{
		db:  db,
		rdx: rdx,
	}
	return
}

func (q *Quarry) Dig(query Address) (m Match) {
	m.Query = q.db.enrichPostcode(query)
	maxdist := 0.0
	for addr := range q.db.Search(m.Query) {
		d := m.Query.dist(addr)
		if d > maxdist {
			maxdist = d
			m.Result = addr
		}
		//util.Info("%s: %f", addr, d)
	}
	if maxdist == 0 {
		m.Result = m.Query
		m.Result.HouseNumber = ""
		m.Result.Street = ""
	} else {
		m.Result.City = m.Query.City
		m.Result.Region = m.Query.Region
		m.Result.Country = m.Query.Country
		m.Result.Postcode = m.Query.Postcode
	}
	return m
}

func (q *Quarry) DigFeed(feed <-chan Address) <-chan Match {
	matchs := make(chan Match)
	wg := &sync.WaitGroup{}
	for i := 0; i < workers(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for addr := range feed {
				matchs <- q.Dig(addr)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(matchs)
	}()
	return matchs
}

func (q *Quarry) CsvFeed(path, col string, delim rune) {
	queries := csvFeed(path, col, ',')
	matchs := q.DigFeed(queries)
	for match := range matchs {
		util.Println(match.String())
	}
}

func (q *Quarry) Excavate(pbf, postcodes string) (err error) {
	util.Info("Excavating...")
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go q.survey(postcodes, wg)
	go q.excavate(pbf, workers(), wg)
	wg.Wait()
	q.index()
	return
}

func (q *Quarry) index() {
	q.rdx = loadRIndex(q.db)
	q.db.indexAddresses(q.rdx)
}

func (q *Quarry) survey(postcode_path string, wg *sync.WaitGroup) {
	defer wg.Done()
	postcodes := ReadPostcodes(postcode_path)
	recs := make(chan QdbRecord)
	go func() {
		defer close(recs)
		for p := range postcodes {
			p.RelationKey = q.rdx.getRelation(p.Center)
			recs <- p
		}
	}()
	q.db.addRecords(PostcodeBucket, recs)
}

func (q *Quarry) excavate(pbf string, workers int, wg *sync.WaitGroup) {
	defer wg.Done()
	filter := NewOsmFilter(1 << 27)
	q.readRelations(pbf, filter, workers)
	q.markNodes(pbf, filter, workers)
	q.addNodes(pbf, filter, 1) //Nodes need to be added sequentially
}

func (q *Quarry) readRelations(pbf string, addrFilter *OsmFilter, workers int) {
	util.Info("Reading Relations...")
	defer util.Info("Done Reading Relations")
	ex, err := osm.NewExcavator(pbf)
	util.Check(err)
	ex.RelationCourier = func(feed <-chan *osm.Relation) {
		rels := make(chan QdbRecord)
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
		q.db.addRecords(RelationBucket, rels)
	}
	err = ex.Start(workers)
	util.Check(err)
}

func (q *Quarry) markNodes(pbf string, addrFilter *OsmFilter, workers int) {
	util.Info("Marking Nodes...")
	defer util.Info("Done Marking Nodes")
	ex, err := osm.NewExcavator(pbf)
	util.Check(err)
	ex.WayCourier = func(feed <-chan *osm.Way) {
		ways := make(chan QdbRecord)
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
		q.db.addRecords(WayBucket, ways)
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
}

// Workers should be set to 1. Qdb wants sequential IDs
func (q *Quarry) addNodes(pbf string, addrFilter *OsmFilter, workers int) {
	util.Info("Adding Nodes...")
	util.Info("Done Adding Nodes")
	ex, err := osm.NewExcavator(pbf)
	util.Check(err)
	ex.NodeCourier = func(feed <-chan *osm.Node) {
		nodes := make(chan QdbRecord)
		go func() {
			defer close(nodes)
			for node := range feed {
				if addrFilter.HasInt64(node.ID) {
					nodes <- node
				}
			}
		}()
		q.db.addRecords(NodeBucket, nodes)
	}
	err = ex.Start(workers)
	util.Check(err)
}

func addressRelations(q *Qdb, addr Address) <-chan string {
	keys := make(chan string)
	go func() {
		defer close(keys)
		for pc := range q.Postcodes(addr.Country, addr.Postcode) {
			keys <- pc.RelationKey
		}
	}()
	return keys
}
