package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"strings"
)

type Address struct {
	ID          int64
	Location    geo.Coordinate
	Street      string
	HouseNumber string
	Postcode    string
	City        string
	Region      string
	Country     string
	Iso2        string
	Tags        map[string]string
}

func NodeAddress(node *osm.Node) Address {
	hn := node.Tags[osm.AddrHouseNum]
	st := node.Tags[osm.AddrStreet]
	if st == "" {
		st = node.Tags["name"]
	}
	pc := node.Tags[osm.AddrPostcode]
	tags := node.Tags
	loc := geo.Coordinate{Lat: node.Lat, Lon: node.Lon}
	return Address{ID: node.ID, HouseNumber: hn, Street: st, Postcode: pc, Tags: tags, Location: loc}
}

func QueryAddress(query string) Address {
	tags := make(map[string]string)
	params := strings.Split(query, "&")
	for _, param := range params {
		kv := strings.Split(param, "=")
		tags[kv[0]] = kv[1]
	}
	return Address{HouseNumber: tags["house"], Street: tags["street"], Postcode: tags["postcode"], Tags: tags}
}

func (a Address) Indexes() <-chan string {
	return mphones(a.Street)
}

func (a Address) dist(to Address) float64 {
	e := editDist(a.Street, a.HouseNumber, to.Street, to.HouseNumber)
	rad := 100000.0 //100km
	d := (rad - a.Location.Distance(to.Location)) / rad
	if d < 0 {
		d = 0
	}
	e += d
	return e
}

func (a Address) String() string {
	addr := []string{a.HouseNumber, a.Street, a.City, a.Postcode, a.Region, a.Country}
	return strings.Join(addr, ", ")
}
