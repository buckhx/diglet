package dig

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
	"strconv"
)

var (
	AddressBucket  = []byte("address")
	NodeBucket     = []byte("node")
	PostcodeBucket = []byte("postcode")
	RelationBucket = []byte("relation")
	WayBucket      = []byte("way")
	empty          struct{}
)

type QdbRecord interface {
	Keyed() (key, value []byte)
	Key() string
	//String() string
}

type Qdb struct {
	path string
	db   *bolt.DB
}

func OpenQdb(path string) (qdb *Qdb, err error) {
	//util.DEBUG = true
	//opts := nil                            //&bolt.Options{}                //MmapFlags: syscall.MAP_POPULATE}
	db, err := bolt.Open(path, 0600, nil) //opts) //&bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		//we're shadowing errors
		//tx.DeleteBucket(PostcodeBucket)
		_, err = tx.CreateBucketIfNotExists(AddressBucket)
		_, err = tx.CreateBucketIfNotExists(PostcodeBucket)
		_, err = tx.CreateBucketIfNotExists(NodeBucket)
		_, err = tx.CreateBucketIfNotExists(RelationBucket)
		_, err = tx.CreateBucketIfNotExists(WayBucket)
		return err
	}); err != nil {
		return
	}
	qdb = &Qdb{
		path: path,
		db:   db,
	}
	return
}

func (q *Qdb) Search(query Address) <-chan Address {
	indexes := query.Indexes()
	addresses := make(chan Address)
	go func() {
		defer close(addresses)
		q.db.View(func(tx *bolt.Tx) error {
			ab := tx.Bucket(AddressBucket)
			nb := tx.Bucket(NodeBucket)
			wb := tx.Bucket(WayBucket)
			for rkey := range addressRelations(q, query) {
				for idx := range indexes {
					idx := util.Sprintf("%s:%s", rkey, idx)
					//util.Info("idx: %s", idx)
					c := ab.Cursor()
					pre := []byte(idx)
					for k, v := c.Seek(pre); bytes.HasPrefix(k, pre); k, v = c.Next() {
						nids := osm.UnmarshalNids(v)
						for _, nid := range nids {
							var node *osm.Node
							//util.Info("%d", nid)
							if nid < 0 {
								wid, _ := osm.MarshalID(-1 * nid)
								way, _ := osm.UnmarshalWay(wb.Get(wid))
								id, _ := osm.MarshalID(way.NodeIDs[0])
								node, _ = osm.UnmarshalNode(nb.Get(id))
								node.Tags = way.Tags
							} else {
								id, _ := osm.MarshalID(nid)
								node, _ = osm.UnmarshalNode(nb.Get(id))
							}
							addresses <- NodeAddress(node)
						}
					}
				}
			}
			return nil
		})
	}()
	return addresses
}

func (q *Qdb) Nodes(tags ...string) <-chan *osm.Node {
	nodes := make(chan *osm.Node, 1<<15)
	go func() {
		defer close(nodes)
		err := q.db.View(func(tx *bolt.Tx) error {
			err := tx.Bucket(NodeBucket).ForEach(func(k, v []byte) error {
				node, err := osm.UnmarshalNode(v)
				if err != nil {
					return err
				}
				if len(tags) == 0 { //no filter
					nodes <- node
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

func (q *Qdb) Relations() <-chan *osm.Relation {
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

func (q *Qdb) WayIDs(ids ...int64) <-chan *osm.Way {
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

func (q *Qdb) NodeIDs(ids ...int64) <-chan *osm.Node {
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

func (q *Qdb) AddressableNodes() <-chan *osm.Node {
	addrs := make(chan *osm.Node)
	go func() {
		defer close(addrs)
		err := q.db.View(func(tx *bolt.Tx) error {
			nb := tx.Bucket(NodeBucket)
			wb := tx.Bucket(WayBucket)
			err := nb.ForEach(func(k, v []byte) error {
				node, err := osm.UnmarshalNode(v)
				if err != nil {
					return err
				}
				if node.IsAddressable() {
					addrs <- node
				}
				return nil
			})
			if err != nil {
				return err
			}
			err = wb.ForEach(func(k, v []byte) error {
				way, err := osm.UnmarshalWay(v)
				if err != nil {
					return err
				}
				if way.IsAddressable() {
					nid, err := osm.MarshalID(way.NodeIDs[0])
					if err != nil {
						return err
					}
					v := nb.Get(nid)
					node, err := osm.UnmarshalNode(v)
					if err != nil {
						return err
					}
					node.Tags = way.Tags
					node.Tags["node_id"] = strconv.FormatInt(node.ID, 10)
					node.Tags["way_id"] = strconv.FormatInt(way.ID, 10)
					node.ID = -1 * way.ID
					addrs <- node
				}
				return nil
			})
			return err
		})
		util.Check(err)

	}()
	return addrs
}

func (q *Qdb) WayNodes(wid int64) (way *osm.Way, nodes []*osm.Node) {
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

//TODO move to quarry
func (q *Qdb) indexAddresses(rdx *rIndex) error {
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
	for node := range q.AddressableNodes() {
		addr := NodeAddress(node)
		rkey := rdx.getNodeRelation(node)
		for idx := range addr.Indexes() {
			idx = util.Sprintf("%s:%s", rkey, idx)
			indexes[idx] = append(indexes[idx], addr.ID) //TODO unique
		}
		if i%batchSize == 0 && i > 0 {
			err := flushIndexes(indexes)
			util.Check(err)
			/*
				if err != nil {
					return err
				}
			*/
			util.Info("Flushing %d addresses @ %d", len(indexes), i)
			indexes = make(map[string][]int64, batchSize)
		}
		i++
	}
	err := flushIndexes(indexes)
	util.Info("Indexed %d addresses", i)
	return err
}

func (q *Qdb) AddRelations(relations <-chan *osm.Relation) {
	elems := make(chan QdbRecord)
	go func() {
		defer close(elems)
		for e := range relations {
			elems <- e
		}
	}()
	q.addRecords(RelationBucket, elems)
}

func (q *Qdb) AddWays(ways <-chan *osm.Way) {
	recs := make(chan QdbRecord)
	go func() {
		defer close(recs)
		for w := range ways {
			recs <- w
		}
	}()
	q.addRecords(WayBucket, recs)
}

func (q *Qdb) AddNodes(nodes <-chan *osm.Node) {
	recs := make(chan QdbRecord)
	go func() {
		defer close(recs)
		for node := range nodes {
			// TODO filter has street or in bloom
			recs <- node
		}
	}()
	q.addRecords(NodeBucket, recs)
}

// Get postcodes for a country. Uses a prefix so ("US", "") will get all US postcodes
func (q *Qdb) Postcodes(countrycode, postcode string) <-chan *Postcode {
	posts := make(chan *Postcode, 1<<10)
	go func() {
		defer close(posts)
		q.db.View(func(tx *bolt.Tx) error {
			key := []byte(util.Sprintf("%s:%s", countrycode, postcode))
			c := tx.Bucket(PostcodeBucket).Cursor()
			for k, v := c.Seek(key); bytes.HasPrefix(k, key); k, v = c.Next() {
				pc, err := unmarshalPostcode(v)
				util.Check(err)
				posts <- pc
				return nil
			}
			return nil
		})
	}()
	return posts
}

func (q *Qdb) enrichPostcode(addr Address) Address {
	var pc *Postcode
	for p := range q.Postcodes(addr.Country, addr.Postcode) {
		pc = p
	}
	if pc == nil {
		return addr
	}
	addr.City = pc.PlaceName
	addr.Region = pc.RegionCode
	addr.Location = pc.Center
	return addr
}

func (q *Qdb) addRecords(bucket []byte, recs <-chan QdbRecord) {
	i := 0
	capacity := 1 << 16 //BlockSize
	batch := make([]QdbRecord, capacity)
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
func (q *Qdb) flush(bucket []byte, recs []QdbRecord) error {
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

func (q *Qdb) Close() {
	q.db.Close()
}
