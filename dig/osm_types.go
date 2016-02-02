package dig

import (
	"github.com/antzucaro/matchr"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
)

const (
	AddrHouseNum = "addr:housenumber"
	AddrStreet   = "addr:street"
	AddrCity     = "addr:city"
	AddrCountry  = "addr:country"

	AddrPrefix  = "addr:"
	GnisPrefix  = "gnis:"
	TigerPrefix = "tiger:"
)

type Node struct {
	*osmpbf.Node
}

type Way struct {
	*osmpbf.Way
}

type Relationstruct struct {
	*osmpbf.Relation
}
