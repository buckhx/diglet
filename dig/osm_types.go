package dig

import (
	_ "bytes"
	_ "encoding/gob"
	_ "github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	AddrHouseNum = "addr:housenumber"
	AddrStreet   = "addr:street"
	AddrCity     = "addr:city"
	AddrCountry  = "addr:country"

	AddrPrefix  = "addr:"
	GnisPrefix  = "gnis:"
	TigerPrefix = "tiger:"
	BlockSize   = 8000
)

type OsmElement interface {
	Keyed() (key, value []byte)
	Valid() bool
	GetID() int64
}

func addressNodes(addr string, nodeIDs []int64) (k, v []byte) {
	k, err := msgpack.Marshal(addr)
	if err != nil {
		return
	}
	v, err = msgpack.Marshal(nodeIDs)
	if err != nil {
		k = nil
	}
	return
}

func unmarshalNids(b []byte) (nids []int64) {
	_ = msgpack.Unmarshal(b, &nids)
	return
}

func nodeAddrs(node *Node) (addrs <-chan string) {
	// metaphone terms
	return mphones(node.Tags[AddrStreet])
}

type AddressedNode struct {
	Key     string
	NodeIDs []int64
}

type Node struct {
	*osmpbf.Node
}

func unmarshalNode(b []byte) (o *Node, err error) {
	err = msgpack.Unmarshal(b, &o)
	return
}

func (o *Node) GetID() int64 {
	return o.ID
}

func (o *Node) Keyed() (k, v []byte) {
	k, err := msgpack.Marshal(o.ID)
	if err != nil {
		return
	}
	v, err = msgpack.Marshal(o)
	if err != nil {
		k = nil
	}
	return
}

func (o *Node) Valid() bool {
	return o.Info.Visible
}

type Way struct {
	*osmpbf.Way
}

func unmarshalWay(b []byte) (o *Way, err error) {
	err = msgpack.Unmarshal(b, &o)
	return
}

func (o *Way) GetID() int64 {
	return o.ID
}

func (o *Way) Keyed() (k, v []byte) {
	k, err := msgpack.Marshal(o.ID)
	if err != nil {
		return
	}
	v, err = msgpack.Marshal(o)
	if err != nil {
		k = nil
	}
	return
}

func (o *Way) Valid() bool {
	return o.Info.Visible
}

type Relation struct {
	*osmpbf.Relation
}

func marshalID(id int64) ([]byte, error) {
	return msgpack.Marshal(id)
}
