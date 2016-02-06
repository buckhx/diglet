package dig

import (
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/util"
)

var (
	AddressBucket  = []byte("address")
	NodeBucket     = []byte("node")
	WayBucket      = []byte("way")
	RelationBucket = []byte("relation")
	empty          struct{}
)

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
		//we're shadowing one of them here
		_, err = tx.CreateBucketIfNotExists(AddressBucket)
		_, err = tx.CreateBucketIfNotExists(NodeBucket)
		_, err = tx.CreateBucketIfNotExists(WayBucket)
		_, err = tx.CreateBucketIfNotExists(RelationBucket)
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

func (q *Quarry) Dig(house, street string) {
	indexes := mphones(street)
	err := q.db.View(func(tx *bolt.Tx) error {
		ab := tx.Bucket(AddressBucket)
		nb := tx.Bucket(NodeBucket)
		for idx := range indexes {
			k, _ := addressNodes(idx, nil)
			nids := unmarshalNids(ab.Get(k))
			for _, nid := range nids {
				id, _ := marshalID(nid)
				node, _ := unmarshalNode(nb.Get(id))
				util.Info("%s", node)
			}
		}
		return nil
	})
	_ = err
}

func (q *Quarry) indexAddresses() error {
	util.Info("Indexing addresses...")
	defer util.Info("Done indexing addresses")
	err := q.db.Update(func(tx *bolt.Tx) error {
		ab := tx.Bucket(AddressBucket)
		addresses := make(map[string][]int64)
		err := tx.Bucket(NodeBucket).ForEach(func(k, v []byte) error {
			node, err := unmarshalNode(v)
			if err != nil {
				return err
			}
			addrs := nodeAddrs(node)
			for addr := range addrs {
				addresses[addr] = append(addresses[addr], node.ID)
			}
			return nil
		})
		for addr, ids := range addresses {
			util.Info("a: %s n: %d", addr, len(ids))
			k, v := addressNodes(addr, ids)
			err = ab.Put(k, v)
			if err != nil {
				return err
			}
		}
		return err
	})
	return err
}

func (q *Quarry) Excavate(pbf string) error {
	util.Info("Loading Primitives")
	defer util.Info("Done loading primitives")
	//q.db.NoSync = true
	ex, err := NewExcavator(pbf)
	if err != nil {
		return err
	}
	ex.WayCourier = q.AddWays
	ex.NodeCourier = q.AddNodes
	err = ex.Start(1)
	if err != nil {
		return err
	}
	util.Info("Enriching nodes")
	err = q.db.Update(func(tx *bolt.Tx) error {
		enriched := 0
		nb := tx.Bucket(NodeBucket)
		err := tx.Bucket(WayBucket).ForEach(func(k, v []byte) error {
			way, err := unmarshalWay(v)
			if err != nil {
				return err
			}
			if _, ok := way.Tags[AddrStreet]; ok {
				nid, err := marshalID(way.NodeIDs[0])
				if err != nil {
					return err
				}
				node, err := unmarshalNode(nb.Get(nid))
				if err != nil {
					return err
				}
				node.Tags[AddrHouseNum] = way.Tags[AddrHouseNum]
				node.Tags[AddrStreet] = way.Tags[AddrStreet]
				k, v = node.Keyed()
				err = nb.Put(k, v)
				if err != nil {
					return err
				}
				enriched++
			}
			return nil
		})
		util.Info("Enriched %d nodes", enriched)
		return err
	})
	if err != nil {
		return err
	}
	err = q.indexAddresses()
	/*
		ex.WayCourier = func(ways <-chan *Way) {
			util.Info("IN COURIER")
			joins := make(chan *Way)
			defer close(joins)
			//	go q.AddWays(joins)
			for way := range ways {
				util.Info("WAY: %+v", way)
				if _, ok := way.Tags[AddrHouseNum]; ok {
					way.NodeIDs = way.NodeIDs[:1]
					way.ID = way.NodeIDs[0]
					joins <- way
				}
			}
		}
		err = ex.Start(1)
		if err != nil {
			return err
		}
		ex.NodeCourier = func(nodes <-chan *Node) {
			inserts := make(chan *Node)
			q.AddNodes(inserts)
			for node := range nodes {
				if _, ok := node.Tags[AddrStreet]; ok {
					inserts <- node
				}
			}
		}
		err = ex.Restart(1)
		if err != nil {
			return err
		}
	*/
	return nil
}

func (q *Quarry) AddWays(ways <-chan *Way) {
	i := 0
	capacity := 65536 //BlockSize
	batch := make([]OsmElement, capacity)
	for way := range ways {
		batch[i] = way
		if i >= capacity-1 {
			err := q.flush(WayBucket, batch)
			util.Warn(err, "batch error")
			i = -1
		}
		i++
	}
	err := q.flush(WayBucket, batch[:i+1])
	util.Warn(err, "batch error")
}

func (q *Quarry) AddNodes(nodes <-chan *Node) {
	i := 0
	capacity := 65536 //BlockSize
	batch := make([]OsmElement, capacity)
	for node := range nodes {
		n := i % capacity
		batch[n] = node
		if n == 0 && i > 0 {
			err := q.flush(NodeBucket, batch)
			util.Warn(err, "batch error")
		}
		i++
	}
	err := q.flush(NodeBucket, batch[:i%capacity])
	util.Warn(err, "batch error")
	util.Info("Added %d nodes", i)
}

func (q *Quarry) PrintStats() {
	q.db.View(func(tx *bolt.Tx) error {
		util.Info("Print node stats")
		util.Info("%+v", tx.Bucket(NodeBucket).Stats())
		return nil
	})
}

// Write an ordered key
func (q *Quarry) flush(bucket []byte, elements []OsmElement) error {
	util.Info("Flushing batch %s %d -> %d", bucket, elements[0].GetID(), elements[len(elements)-1].GetID())
	return q.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		for _, elem := range elements {
			k, v := elem.Keyed()
			err := b.Put(k, v)
			if err != nil {
				return err
			}
		}
		util.Debug("TX - id: %d, stats: %+v", tx.ID, tx.Stats())
		return nil
	})
}

func (q *Quarry) Close() {
	q.db.Close()
}
