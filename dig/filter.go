package dig

import (
	"encoding/binary"
	"github.com/buckhx/diglet/util"
	"github.com/tylertreat/boomfilters"
)

const (
	fpRate = 0.01
)

type OsmFilter struct {
	filter boom.Filter
	//filter *boom.CuckooFilter
}

func NewOsmFilter(filesize int64) (o *OsmFilter) {
	n := filesize >> 7
	util.Info("%d", n)
	o = &OsmFilter{
		filter: boom.NewScalableBloomFilter(uint(n), fpRate, 0.9),
		//filter: boom.NewCuckooFilter(uint(n), fpRate),
	}
	return
}

func (o *OsmFilter) Add(k []byte) *OsmFilter {
	o.filter.Add(k)
	return o
}

func (o *OsmFilter) Has(k []byte) bool {
	return o.filter.Test(k)
}

func (o *OsmFilter) AddInt64(k int64) *OsmFilter {
	b := make([]byte, 8) //slow?
	binary.PutVarint(b, k)
	return o.Add(b)
}

func (o *OsmFilter) HasInt64(k int64) bool {
	b := make([]byte, 8) //slow?
	binary.PutVarint(b, k)
	return o.Has(b)
}
