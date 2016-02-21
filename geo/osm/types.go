package osm

import (
	_ "bytes"
	_ "encoding/gob"
	"github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
	"gopkg.in/vmihailenco/msgpack.v2"
	"strconv"
)

//TODO split types into their own module
const (
	AddrHouseNum = "addr:housenumber"
	AddrStreet   = "addr:street"
	AddrCity     = "addr:city"
	AddrCountry  = "addr:country"
	AddrPostcode = "addr:postcode"
	AdminLevel   = "admin_level"
	Boundary     = "boundary"

	AddrPrefix  = "addr:"
	GnisPrefix  = "gnis:"
	TigerPrefix = "tiger:"
	BlockSize   = 8000

	RoleOuter = "outer"
	RoleInner = "inner"
)

func MarshalAddrIndex(idx string, nodeIDs []int64) (k, v []byte) {
	k = []byte(idx)
	v, err := msgpack.Marshal(nodeIDs)
	if err != nil {
		k = nil
	}
	return
}

func UnmarshalNids(b []byte) (nids []int64) {
	_ = msgpack.Unmarshal(b, &nids)
	return
}

type Node struct {
	*osmpbf.Node
}

func UnmarshalNode(b []byte) (o *Node, err error) {
	err = msgpack.Unmarshal(b, &o)
	return
}

func (o *Node) IsAddressable() bool {
	return o.Valid() && o.Tags[AddrStreet] != ""
}

func (o *Node) Key() string {
	return strconv.FormatInt(o.ID, 10)
}

func (o *Node) String() string {
	return util.Sprintf("%+v", o)
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

func (o *Node) GetID() int64 {
	return o.ID
}

func (o *Node) Valid() bool {
	return o.Info.Visible
}

type Way struct {
	*osmpbf.Way
}

func (o *Way) IsSubregionBoundary() bool {
	return o.Tags[AdminLevel] == "6" && o.Tags[Boundary] == "administrative"
}

func (o *Way) IsAddressable() bool {
	return o.Valid() && (o.Tags[AddrStreet] != "" || o.Tags["highway"] == "residential")
}

func UnmarshalWay(b []byte) (o *Way, err error) {
	err = msgpack.Unmarshal(b, &o)
	return
}

func (o *Way) Key() string {
	return strconv.FormatInt(o.ID, 10)
}

func (o *Way) String() string {
	return util.Sprintf("%+v", o)
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

func UnmarshalRelation(b []byte) (o *Relation, err error) {
	err = msgpack.Unmarshal(b, &o)
	return
}

func (o *Relation) IsSubregionBoundary() bool {
	return o.Tags[AdminLevel] == "6" && o.Tags[Boundary] == "administrative"
}

func (o *Relation) Key() string {
	return strconv.FormatInt(o.ID, 10)
}

func (o *Relation) String() string {
	return util.Sprintf("%+v", o)
}

func (o *Relation) Keyed() (k, v []byte) {
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

func (o *Relation) Valid() bool {
	return o.Info.Visible
}

func MarshalID(id int64) ([]byte, error) {
	return msgpack.Marshal(id)
}

var (
	NodeType     = osmpbf.NodeType
	WayType      = osmpbf.WayType
	RelationType = osmpbf.RelationType
)
