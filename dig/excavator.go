package dig

import (
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
)

type Excavator struct {
	pbfFile   *File
	db        *bolt.DB
	nodes     chan Node
	ways      chan Way
	relations chan Relation
	diggers   int
}

func NewExcavator(dbPath, pbfPath string) (ex *Excavator, err error) {
	pbf, err := os.Open(pbfPath)
	if err != nil {
		return
	}
	db, err := bolt.Open(dbPath, 0600, nil) //&bolt.Options{Timeout: 5 * time.Second})
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

func (d *Excavator) Close() {
	d.db.Close()
	d.pbfFile.Close()
}
