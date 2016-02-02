package dig

import (
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/util"
)

type DigDB struct {
	path string
	db   *bolt.DB
}

func NewQuarry(path string) (quarry *Quarry, err error) {
	db, err := bolt.Open(path, 0600, nil) //&bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		//we're shadowing one of them here
		_, err = tx.CreateBucketIfNotExists(AddressBucket)
		_, err = tx.CreateBucketIfNotExists(NodeBucket)
		return err
	}); err != nil {
		return
	}

}

func (d *Digger) Close() {
	d.db.Close()
	d.pbfFile.Close()
}
