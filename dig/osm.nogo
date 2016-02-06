package dig

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/antzucaro/matchr"
	"github.com/boltdb/bolt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const DEBUG = false
const (
	ZoomRes   = 23
	BlockSize = 8000
)

var (
	AddressBucket = []byte("address")
	NodeBucket    = []byte("node")
	empty         struct{}
	threads       = 1
)

type nodeset map[int64]struct{}

func (set nodeset) marshal() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(set)
	return buf.Bytes(), err
}

func unmarshalNodeset(raw []byte) (set nodeset, err error) {
	err = gob.NewDecoder(bytes.NewReader(raw)).Decode(&set)
	return
}

func encoder(db *bolt.DB, nodes <-chan *osmpbf.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	size := 1000
	i := 0
	flushed := 0
	batch := make([]*osmpbf.Node, size)
	flush := func() error {
		util.Info("Flushing batch %d", flushed)
		return db.Batch(func(tx *bolt.Tx) (err error) {
			ab := tx.Bucket(AddressBucket)
			nb := tx.Bucket(NodeBucket)
			nodes := make(map[string]nodeset, size)
			for _, node := range batch {
				for _, mp := range strings.Split(node.Tags["dig:key"], ",") {
					if group, ok := nodes[mp]; !ok {
						if g := ab.Get([]byte(mp)); g == nil {
							group = make(nodeset)
						} else if group, err = unmarshalNodeset(g); err != nil {
							return
						}
						nodes[mp] = group
					}
					nodes[mp][node.ID] = empty
				}
				// Insert node
				v, err := marshalNode(node)
				k := marshalId(node.ID)
				if err = nb.Put(k, v); err != nil {
					return err
				}
			}
			for mp, group := range nodes {
				if value, err := group.marshal(); err != nil {
					return err
				} else {
					if err = ab.Put([]byte(mp), value); err != nil {
						return err
					}
				}
			}
			return nil
		})
	}
	for node := range nodes {
		batch[i] = node
		if i >= size-1 {
			flushed++
			err := flush()
			util.Warn(err, "batch error")
			i = -1
		}
		i++
	}
	err := flush()
	util.Warn(err, "batch error")
}

func decoder(pbf *osmpbf.Decoder, nodes chan<- *osmpbf.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if v, err := pbf.Decode(); err == io.EOF {
			return
		} else if err != nil {
			util.Fatal("decode error %v", err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				if hasPrefix(v.Tags, AddrStreet) {
					tagQuadkey(v)
					nodes <- v
					if strings.Contains(v.Tags[AddrStreet], " 42nd") {
						util.Info("%+v", v)
					}
				}
			case *osmpbf.Way:
				//util.Info("Way: %+v", v)
			case *osmpbf.Relation:
				/*
					if hasPrefix(v.Tags, "ISO") {
						//if hasTag(v.Tags, "boundary", "administrative") && hasTag(v.Tags, "admin_level", "6") {
						util.Info("Relation: %+v", v)
					}
				*/
			default:
				util.Fatal("unknown type %T", v)
			}
		}
	}
}

func hasPrefix(tags map[string]string, prefix string) bool {
	for k := range tags {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func hasTag(tags map[string]string, key, value string) bool {
	for k, v := range tags {
		if k == key && v == value {
			return true
		}
	}
	return false
}

func tagQuadkey(node *osmpbf.Node) {
	tile, _ := tile_system.CoordinateToTile(node.Lat, node.Lon, 23)
	lat := strconv.FormatFloat(node.Lat, 'f', 8, 64)
	lon := strconv.FormatFloat(node.Lon, 'f', 8, 64)
	node.Tags["dig:lat"] = lat
	node.Tags["dig:lon"] = lon
	node.Tags["dig:key"] = nodeKey(node)
	node.Tags["dig:qk"] = tile.QuadKey()
}

func key(value string) string {
	var b1, b2 bytes.Buffer
	value = expand(clean(value))
	terms := strings.Split(value, " ")
	for _, term := range terms {
		m1, m2 := matchr.DoubleMetaphone(term)
		b1.WriteString(m1)
		b2.WriteString(m2)
	}
	b1.WriteString(",")
	b1.Write(b2.Bytes())
	return b1.String()
}

var nonword = regexp.MustCompile("[^\\w ]")
var expansions = map[string]string{
	"0":     "zero ",
	"1":     "one ",
	"2":     "two ",
	"3":     "three ",
	"4":     "four ",
	"5":     "five ",
	"6":     "six ",
	"7":     "seven ",
	"8":     "eight ",
	"9":     "nine ",
	" n ":   "north ",
	" e ":   "este ",
	" s ":   "south ",
	" w ":   "oost ",
	"north": "north ", //for
	"east":  "este ",
	"south": "south ",
	"west":  "oost ",
}

func clean(s string) string {
	s = strings.ToLower(s)
	return nonword.ReplaceAllString(s, "")
}

func expand(s string) string {
	for i, o := range expansions {
		s = strings.Replace(s, i, o, -1)
	}
	s = strings.Replace(s, "  ", " ", -1)
	return util.Sprintf(" %s ", s)
}

func marshalId(id int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, id)
	return b
}

func unmarshalId(b []byte) (id int64) {
	id, _ = binary.Varint(b)
	return
}

func marshalNode(node *osmpbf.Node) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(node)
	return buf.Bytes(), err
}

func unmarshalNode(raw []byte) (node *osmpbf.Node, err error) {
	err = gob.NewDecoder(bytes.NewReader(raw)).Decode(&node)
	return
}

func nodeKey(node *osmpbf.Node) string {
	//k := strings.Join([]string{node.Tags[AddrHouseNum], node.Tags[AddrStreet]}, " ")
	k := node.Tags[AddrStreet]
	k = key(k)
	return k
}

func queryKeys(q string) []string {
	util.Info(q)
	tags := strings.Split(q, " ")
	util.Info("%q", tags)
	k := tags[1] // skip housenumber
	util.Info(k)
	c := string(k[0]) //check first char for expansions
	if _, ok := expansions[k]; ok {
		k += tags[2] // include the second tag as part of key
	} else if _, ok := expansions[c]; ok {
		k += tags[2] // include the second tag as part of key
	}
	util.Info(k)
	k = expand(k)
	util.Info(k)
	k = key(k)
	ks := strings.Split(k, ",")
	util.Info("%s", ks)
	return ks
}
