package mbt

import (
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	ts "github.com/buckhx/tiles"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	featureBucket = []byte("feature")
)

type tileFeatures struct {
	t ts.Tile
	f []*geo.Feature
}
type fID []byte

type featureIndex struct {
	path string
	db   *bolt.DB
	idx  ts.TileIndex
}

func newFeatureCache(path string) (c *featureIndex, err error) {
	// TODO rm path first just to make sure we don't reuse any data
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(featureBucket)
		return err
	}); err != nil {
		return
	}
	c = &featureIndex{path: path, db: db, idx: ts.NewTileIndex()}
	return
}

func (c *featureIndex) tileFeatures(zmin, zmax int) <-chan tileFeatures {
	tfs := make(chan tileFeatures, 1<<10)
	go func() {
		defer close(tfs)
		for t := range c.idx.TileRange(zmin, zmax) {
			fids := c.idx.Values(t)
			c.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(featureBucket)
				tf := tileFeatures{t: t, f: make([]*geo.Feature, len(fids))}
				for i, f := range fids {
					k := f.(fID)
					v := b.Get(k)
					err := msgpack.Unmarshal(v, &tf.f[i])
					util.Check(err)
				}
				tfs <- tf
				return nil
			})
		}
	}()
	return tfs
}

func (c *featureIndex) indexFeatures(features <-chan *geo.Feature) {
	records := make(chan *geo.Feature, 1<<10)
	go func() {
		defer close(records)
		for f := range features {
			k, err := key(f)
			if err != nil {
				util.Warn(err, "key err")
				continue
			}
			records <- f
			//TODO parallelize
			for _, t := range FeatureTiles(f) {
				c.idx.Add(t, k)
			}
		}
	}()
	c.addRecords(featureBucket, records)
}

func (c *featureIndex) addRecords(bucket []byte, recs <-chan *geo.Feature) {
	i := 0
	capacity := 1 << 16 //BlockSize
	batch := make([]*geo.Feature, capacity)
	for rec := range recs {
		n := i % capacity
		if n == 0 && i > 0 {
			err := c.flush(bucket, batch)
			util.Warn(err, "batch error")
		}
		batch[n] = rec
		i++
	}
	err := c.flush(bucket, batch[:i%capacity])
	util.Warn(err, "batch error")
	util.Info("Added %d %ss", i, bucket)
}

// Write an ordered key
func (c *featureIndex) flush(bucket []byte, recs []*geo.Feature) error {
	if len(recs) < 1 {
		util.Info("Flushing empty batch, skipping")
		return nil
	}
	util.Info("Flushing batch %s %d -> %d", bucket, recs[0].ID, recs[len(recs)-1].ID)
	return c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		for _, rec := range recs {
			k, v := keyed(rec)
			err := b.Put(k, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *featureIndex) close() {
	c.db.Close()
	os.Remove(c.path)
}

func keyed(f *geo.Feature) (k, v []byte) {
	k, err := msgpack.Marshal(f.ID)
	if err != nil {
		return
	}
	v, err = msgpack.Marshal(f)
	if err != nil {
		k = nil
	}
	return
}

func key(f *geo.Feature) (fID, error) {
	k, err := msgpack.Marshal(f.ID)
	return fID(k), err
}
