package tms

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/buckhx/mbtiles"
)

type TileCache struct {
	*Cache
}

func InitTileCache(path string) (cache *TileCache, err error) {
	c, err := InitCache(path)
	cache = &TileCache{Cache: c}
	return
}

func (c *TileCache) GetTile(tileset, key string) (tile *mbtiles.Tile, ok bool) {
	if raw, ok := c.Get(tileset, key); ok {
		var err error
		tile, err = unmarshalTile(raw)
		ok = err == nil
	}
	return
}

func (c *TileCache) PutTile(tileset, key string, tile *mbtiles.Tile) (err error) {
	raw, err := marshalTile(tile)
	if err == nil {
		err = c.Put(tileset, key, raw)
	}
	return
}

func (c *TileCache) MapTiles(tileset string, fn func(value *mbtiles.Tile) (*mbtiles.Tile, error)) (err error) {
	curry := func(value []byte) ([]byte, error) {
		tile, err := unmarshalTile(value)
		if err != nil {
			return nil, err
		}
		tile, err = fn(tile)
		if err != nil {
			return nil, err
		}
		return marshalTile(tile)
	}
	return c.Map(tileset, curry)
}

func marshalTile(tile *mbtiles.Tile) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(tile)
	return buf.Bytes(), err
}

func unmarshalTile(raw []byte) (tile *mbtiles.Tile, err error) {
	err = gob.NewDecoder(bytes.NewReader(raw)).Decode(&tile)
	return
}

type Cache struct {
	path string
	db   *bolt.DB
}

func InitCache(path string) (cache *Cache, err error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	cache = &Cache{path: path, db: db}
	return
}

func (c *Cache) Get(bucket, key string) (value []byte, ok bool) {
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		ok = b != nil
		if ok {
			value = b.Get([]byte(key))
			ok = value != nil && len(value) > 0
		}
		return nil
	})
	return
}

func (c *Cache) Put(bucket, key string, value []byte) (err error) {
	c.db.Update(func(tx *bolt.Tx) error {
		//b := tx.Bucket([]byte("MyBucket"))
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err == nil && len(value) > 0 {
			err = b.Put([]byte(key), value)
		}
		return err
	})
	return
}

func (c *Cache) Map(bucket string, fn func(value []byte) ([]byte, error)) (err error) {
	c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		cur := b.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			val, err := fn(v)
			if err != nil {
				return err
			}
			err = b.Put(k, val)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (c *Cache) DropBucket(bucket string) (ok bool) {
	c.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(bucket))
		ok = err == nil
		return err
	})
	return
}

func (c *Cache) Close() {
	c.db.Close()
}

func (c *Cache) Destroy() {
	c.db.Close()
	os.Remove(c.path)
}
