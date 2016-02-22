package dig

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
)

var (
	AddressBucket  = []byte("address")
	NodeBucket     = []byte("node")
	PostcodeBucket = []byte("postcode")
	RelationBucket = []byte("relation")
	WayBucket      = []byte("way")
	empty          struct{}
)

type QuarryRecord interface {
	Keyed() (key, value []byte)
	Key() string
	String() string
}

type Quarry struct {
	path string
	db   *bolt.DB
}

func NewQuarry(path string) (quarry *Quarry, err error) {
	util.DEBUG = true
	db, err := bolt.Open(path, 0600, nil) //&bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		//we're shadowing errors
		_, err = tx.CreateBucketIfNotExists(AddressBucket)
		_, err = tx.CreateBucketIfNotExists(PostcodeBucket)
		_, err = tx.CreateBucketIfNotExists(NodeBucket)
		_, err = tx.CreateBucketIfNotExists(RelationBucket)
		_, err = tx.CreateBucketIfNotExists(WayBucket)
		return err
	}); err != nil {
		return
	}
	quarry = &Quarry{
		path: path,
		db:   db,
	}
	return
}

func (q *Quarry) Dig(query Address) (match Address) {
	util.Info("QUERY: %s", query)
	indexes := query.Indexes()
	addresses := make(chan Address)
	go func() {
		q.db.View(func(tx *bolt.Tx) error {
			ab := tx.Bucket(AddressBucket)
			nb := tx.Bucket(NodeBucket)
			for idx := range indexes {
				util.Info("%s", idx)
				c := ab.Cursor()
				pre := []byte(idx)
				for k, v := c.Seek(pre); bytes.HasPrefix(k, pre); k, v = c.Next() {
					nids := osm.UnmarshalNids(v)
					for _, nid := range nids {
						id, _ := osm.MarshalID(nid)
						node, _ := osm.UnmarshalNode(nb.Get(id))
						addresses <- NodeAddress(node)
					}
				}
			}
			close(addresses)
			return nil
		})
	}()
	maxdist := 0.0
	for addr := range addresses {
		d := query.edist(addr)
		if d > maxdist {
			maxdist = d
			match = addr
		}
		util.Info("%s: %f", addr, d)
	}
	util.Info("MATCH %s: %f", match, maxdist)
	return
}

func (q *Quarry) Survey(postcode_path string) (err error) {
	postcodes := ReadPostcodes(postcode_path)
	recs := make(chan QuarryRecord)
	go func() {
		defer close(recs)
		for p := range postcodes {
			recs <- p
		}
	}()
	q.addRecords(PostcodeBucket, recs)
	return nil
}

func (q *Quarry) Nodes(tags ...string) <-chan *osm.Node {
	nodes := make(chan *osm.Node, 1<<15)
	go func() {
		defer close(nodes)
		err := q.db.View(func(tx *bolt.Tx) error {
			err := tx.Bucket(NodeBucket).ForEach(func(k, v []byte) error {
				node, err := osm.UnmarshalNode(v)
				if err != nil {
					return err
				}
				for _, tag := range tags {
					if _, ok := node.Tags[tag]; ok {
						nodes <- node
						break
					}
				}
				return nil
			})
			return err
		})
		_ = err
	}()
	return nodes
}

func (q *Quarry) Relations() <-chan *osm.Relation {
	rels := make(chan *osm.Relation, 1<<10)
	go func() {
		defer close(rels)
		err := q.db.View(func(tx *bolt.Tx) error {
			err := tx.Bucket(RelationBucket).ForEach(func(k, v []byte) error {
				rel, err := osm.UnmarshalRelation(v)
				if err != nil {
					return err
				}
				rels <- rel
				return nil
			})
			return err
		})
		util.Check(err)
	}()
	return rels
}

func (q *Quarry) WayIDs(ids ...int64) <-chan *osm.Way {
	ways := make(chan *osm.Way, 1<<10)
	go func() {
		defer close(ways)
		err := q.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(WayBucket)
			for _, id := range ids {
				k, _ := osm.MarshalID(id)
				v := b.Get(k)
				way, err := osm.UnmarshalWay(v)
				if err != nil {
					return err
				}
				ways <- way
			}
			return nil
		})
		util.Check(err)
	}()
	return ways
}

func (q *Quarry) NodeIDs(ids ...int64) <-chan *osm.Node {
	nodes := make(chan *osm.Node, 1<<10)
	go func() {
		defer close(nodes)
		err := q.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(NodeBucket)
			for _, id := range ids {
				k, _ := osm.MarshalID(id)
				v := b.Get(k)
				node, err := osm.UnmarshalNode(v)
				if err != nil {
					return err
				}
				nodes <- node
			}
			return nil
		})
		util.Check(err)
	}()
	return nodes
}

func (q *Quarry) WayNodes(wid int64) (way *osm.Way, nodes []*osm.Node) {
	q.db.View(func(tx *bolt.Tx) error {
		//util.Info("\t%d", wid)
		k, _ := osm.MarshalID(wid)
		w := tx.Bucket(WayBucket).Get(k)
		var err error
		if way, err = osm.UnmarshalWay(w); err != nil {
			return err
		}
		b := tx.Bucket(NodeBucket)
		nodes = make([]*osm.Node, len(way.NodeIDs))
		for i, nid := range way.NodeIDs {
			//util.Info("\t\t%d", nid)
			k, _ := osm.MarshalID(nid)
			v := b.Get(k)
			node, _ := osm.UnmarshalNode(v)
			nodes[i] = node
		}
		return nil
	})
	return
}

func (q *Quarry) indexAddresses() error {
	util.Info("Indexing addresses...")
	defer util.Info("Done indexing addresses")
	flushIndexes := func(indexes map[string][]int64) error {
		err := q.db.Update(func(tx *bolt.Tx) error {
			ab := tx.Bucket(AddressBucket)
			for idx, ids := range indexes {
				k, v := osm.MarshalAddrIndex(idx, ids)
				if val := ab.Get(k); val != nil {
					nids := osm.UnmarshalNids(val)
					ids = append(ids, nids...) //TODO unique
					k, v = osm.MarshalAddrIndex(idx, ids)
				}
				err := ab.Put(k, v)
				if err != nil {
					return err
				}
			}
			return nil
		})
		return err
	}
	i := 0
	batchSize := 1 << 16
	indexes := make(map[string][]int64, batchSize)
	for node := range q.Nodes(osm.AddrStreet) {
		addr := NodeAddress(node)
		for idx := range addr.Indexes() {
			indexes[idx] = append(indexes[idx], node.ID) //TODO unique
		}
		if i%batchSize == 0 && i > 0 {
			err := flushIndexes(indexes)
			util.Info("Flushing %d addresses @ %d", len(indexes), i)
			if err != nil {
				return err
			}
			indexes = make(map[string][]int64, batchSize)
		}
		i++
	}
	err := flushIndexes(indexes)
	util.Info("Indexed %d addresses", i)
	return err
}

func (q *Quarry) AddRelations(relations <-chan *osm.Relation) {
	elems := make(chan QuarryRecord)
	go func() {
		defer close(elems)
		for e := range relations {
			elems <- e
		}
	}()
	q.addRecords(RelationBucket, elems)
}

func (q *Quarry) AddWays(ways <-chan *osm.Way) {
	count := 0
	members := 0
	for way := range ways {
		if _, ok := way.Tags["building"]; ok { // or way.Tags["admin_level"] == "6" or highway {
			//util.Info("%s", way)
			count++
			members += len(way.NodeIDs)
		}
	}
	util.Info("%d ways, %d members", count, members)
	/*
		elems := make(chan OsmElement)
		go func() {
			defer close(elems)
			for e := range ways {
				elems <- e
			}
		}()
		q.addOsmElements(WayBucket, elems)
	*/
}

func (q *Quarry) AddNodes(nodes <-chan *osm.Node) {
	recs := make(chan QuarryRecord)
	go func() {
		defer close(recs)
		for node := range nodes {
			// TODO filter has street or in bloom
			recs <- node
		}
	}()
	q.addRecords(NodeBucket, recs)
}

func (q *Quarry) AddPostcodes(<-chan *Postcode) {

}

func (q *Quarry) addRecords(bucket []byte, recs <-chan QuarryRecord) {
	i := 0
	capacity := 1 << 16 //BlockSize
	batch := make([]QuarryRecord, capacity)
	for rec := range recs {
		n := i % capacity
		if n == 0 && i > 0 {
			err := q.flush(bucket, batch)
			util.Warn(err, "batch error")
		}
		batch[n] = rec
		i++
	}
	err := q.flush(bucket, batch[:i%capacity])
	util.Warn(err, "batch error")
	util.Info("Added %d %ss", i, bucket)
}

// Write an ordered key
func (q *Quarry) flush(bucket []byte, recs []QuarryRecord) error {
	if len(recs) < 1 {
		util.Info("Flushing empty batch, skipping")
		return nil
	}
	util.Info("Flushing batch %s %s -> %s", bucket, recs[0].Key(), recs[len(recs)-1].Key())
	return q.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		for _, rec := range recs {
			k, v := rec.Keyed()
			err := b.Put(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (q *Quarry) Close() {
	q.db.Close()
}
