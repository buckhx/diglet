package dig

import (
	"github.com/buckhx/diglet/geo/osm"
	"strings"
)

type Address struct {
	Latitude    float64
	Longitude   float64
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
	pc := node.Tags[osm.AddrPostcode]
	tags := node.Tags
	return Address{HouseNumber: hn, Street: st, Postcode: pc, Tags: tags}
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
	keys := make(chan string)
	post := a.Postcode
	if len(post) > 5 {
		post = post[:5] // only get first 5 chars of postcode
	}
	go func() {
		defer close(keys)
		for mphone := range mphones(a.Street) {
			keys <- strings.Join([]string{post, mphone}, ":")
		}
	}()
	return keys
}

func (a Address) edist(to Address) float64 {
	return editDist(a.Street, a.HouseNumber, to.Street, to.HouseNumber)
}

func (a Address) String() string {
	addr := []string{a.HouseNumber, a.Street, a.City, a.Region, a.Country, a.Iso2, a.Postcode}
	return strings.Join(addr, ", ")
}
