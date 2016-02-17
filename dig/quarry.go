package dig

import (
	"bytes"
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
		//we're shadowing errors
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
					nids := unmarshalNids(v)
					for _, nid := range nids {
						id, _ := marshalID(nid)
						node, _ := unmarshalNode(nb.Get(id))
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

/*
func (q *Quarry) WayNodes(tags ...string) <-chan *WayNode {
	waynodes := make(chan *WayNode, 1<<15)
	go func() {
		defer close(waynodes)
		prev, _ := marshalID(-1) //first
		for prev != nil {
			q.db.View(func(tx *bolt.Tx) error {
				nb := tx.Bucket(NodeBucket)
				wb := tx.Bucket(WayBucket)
				cur := wb.Cursor()
				cur.Seek(prev)
				for i := 0; i < 1<<15; i++ {
					k, v := cur.Next()
					prev = k
					if k == nil {
						break
					}
					way, err := unmarshalWay(v)
					if err != nil {
						return err
					}
					for _, tag := range tags {
						if _, ok := way.Tags[tag]; ok {
							nodes := make([]*Node, len(way.NodeIDs))
							waynode := &WayNode{Way: way, Nodes: nodes}
							for i, n := range way.NodeIDs {
								nid, err := marshalID(n)
								if err != nil {
									return err
								}
								node, err := unmarshalNode(nb.Get(nid))
								if err != nil {
									return err
								}
								waynode.Nodes[i] = node
							}
							waynodes <- waynode
							break
						}
					}
				}
				return nil
			})
		}
	}()
	return waynodes
}
*/

func (q *Quarry) Nodes(tags ...string) <-chan *Node {
	nodes := make(chan *Node, 1<<15)
	go func() {
		defer close(nodes)
		err := q.db.View(func(tx *bolt.Tx) error {
			err := tx.Bucket(NodeBucket).ForEach(func(k, v []byte) error {
				node, err := unmarshalNode(v)
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

func (q *Quarry) indexAddresses() error {
	util.Info("Indexing addresses...")
	defer util.Info("Done indexing addresses")
	flushIndexes := func(indexes map[string][]int64) error {
		err := q.db.Update(func(tx *bolt.Tx) error {
			ab := tx.Bucket(AddressBucket)
			for idx, ids := range indexes {
				k, v := marshalAddrIndex(idx, ids)
				if val := ab.Get(k); val != nil {
					nids := unmarshalNids(val)
					ids = append(ids, nids...) //TODO unique
					k, v = marshalAddrIndex(idx, ids)
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
	for node := range q.Nodes(AddrStreet) {
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

/*
func (q *Quarry) enrichNodes() error {
	util.Info("Enriching nodes")
	nodes := make(chan *Node, 1<<15)
	defer close(nodes)
	go q.AddNodes(nodes)
	enriched := 0
	for waynode := range q.WayNodes(AddrStreet) {
		enriched++
		way := waynode.Way
		node := waynode.Nodes[0]
		node.Tags[AddrHouseNum] = way.Tags[AddrHouseNum]
		node.Tags[AddrStreet] = way.Tags[AddrStreet]
		nodes <- node
	}
	util.Info("Enriched %d nodes", enriched)
	return nil
}
*/

func (q *Quarry) Excavate(pbf string) (err error) {
	//q.db.NoSync = true
	ex, err := NewExcavator(pbf)
	if err != nil {
		return err
	}
	//ex.NodeCourier = q.AddNodes         //AddressableNodes
	ex.WayCourier = q.AddWays //AddressableNodes
	//ex.RelationCourier = q.AddRelations //AddressableNodes
	err = ex.Start(4)
	if err != nil {
		return err
	}
	/*
		err = q.enrichNodes()
		if err != nil {
			return err
		}
		err = q.indexAddresses()
	*/
	return
}

func (q *Quarry) AddRelations(relations <-chan *Relation) {
	count := 0
	members := 0
	for r := range relations {
		if r.Tags["admin_level"] == "6" {
			util.Info("%s", r)
			count++
			members += len(r.Members)
		}
	}
	util.Info("%d relations, %d members", count, members)
	/*
		elems := make(chan OsmElement)
		go func() {
			defer close(elems)
			for e := range relations {
				elems <- e
			}
		}()
		q.addOsmElements(RelationBucket, elems)
	*/
}

func (q *Quarry) AddWays(ways <-chan *Way) {
	count := 0
	members := 0
	for way := range ways {
		if _, ok := way.Tags["building"]; ok { //way.Tags["admin_level"] == "6" {
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

func (q *Quarry) AddNodes(nodes <-chan *Node) {
	elems := make(chan OsmElement)
	go func() {
		defer close(elems)
		for e := range nodes {
			// TODO filter has street or in bloom
			elems <- e
		}
	}()
	q.addOsmElements(NodeBucket, elems)
}

/*
func (q *Quarry) AddAdminWays(ways <-chan *Way) {
	for way := range ways {
		if adlvl, ok := way.Tags["admin_level"]; ok && adlvl == "6" {
			util.Info("%s %d", way.Tags, len(way.NodeIDs))

			//addrs <- node
		}
	}
}

func (q *Quarry) AddAddressableNodes(nodes <-chan *Node) {
	addrs := make(chan *Node, 1<<16)
	defer close(addrs)
	go q.AddNodes(addrs)
	for node := range nodes {
		if _, ok := node.Tags[AddrStreet]; ok {
			addrs <- node
		}
	}
}
*/

func (q *Quarry) PrintStats() {
	q.db.View(func(tx *bolt.Tx) error {
		util.Info("Print node stats")
		util.Info("%+v", tx.Bucket(NodeBucket).Stats())
		return nil
	})
}

// Write an ordered key
func (q *Quarry) flush(bucket []byte, elements []OsmElement) error {
	if len(elements) < 1 {
		util.Info("Flushing empty batch, skipping")
		return nil
	}
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
		//util.Debug("TX - id: %d, stats: %+v", tx.ID, tx.Stats())
		return nil
	})
}

func (q *Quarry) addOsmElements(bucket []byte, elems <-chan OsmElement) {
	i := 0
	capacity := 1 << 16 //BlockSize
	batch := make([]OsmElement, capacity)
	for e := range elems {
		n := i % capacity
		if n == 0 && i > 0 {
			err := q.flush(bucket, batch)
			util.Warn(err, "batch error")
		}
		batch[n] = e
		i++
	}
	err := q.flush(bucket, batch[:i%capacity])
	util.Warn(err, "batch error")
	util.Info("Added %d %ss", i, bucket)
}

func (q *Quarry) Close() {
	q.db.Close()
}
