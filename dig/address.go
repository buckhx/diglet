package dig

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
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
	return Address{
		ID:          node.ID,
		HouseNumber: hn,
		Street:      st,
		Postcode:    pc,
		Tags:        tags,
		Location:    loc,
	}
}

func QueryAddress(query string) Address {
	tags := make(map[string]string)
	params := strings.Split(query, "&")
	for _, param := range params {
		kv := strings.Split(param, "=")
		tags[kv[0]] = kv[1]
	}
	return Address{
		HouseNumber: tags["house"],
		Street:      tags["street"],
		City:        tags["city"],
		Region:      tags["region"],
		Postcode:    tags["postcode"],
		Country:     tags["country"],
		Tags:        tags,
	}
}

// Expects comma seperated
// house_num,street,city,state,country,zip
func StringAddress(raw string) Address {
	terms := strings.Split(raw, ",")
	addr := Address{}
	for i, v := range terms {
		switch i {
		case 0:
			addr.HouseNumber = v
		case 1:
			addr.Street = v
		case 2:
			addr.City = v
		case 3:
			addr.Region = v
		case 4:
			addr.Country = v
		case 5:
			addr.Postcode = v
		default:
			util.Info("Malformed address %q", raw)
		}
	}
	return addr
}

//Strictly equals hn,st,city,region,country,post
func (a Address) Equals(o Address) bool {
	switch {
	case a.HouseNumber != o.HouseNumber:
		return false
	case a.Street != o.Street:
		return false
	case a.City != o.City:
		return false
	case a.Region != o.Region:
		return false
	case a.Country != o.Country:
		return false
	case a.Postcode != o.Postcode:
		return false
	default:
		return true
	}
}

func (a Address) Indexes() <-chan string {
	return mphones(a.Street)
}

func (a Address) dist(to Address) float64 {
	// TODO break these out into discrete vectors
	e := editDist(a.Street, a.HouseNumber, to.Street, to.HouseNumber)
	rad := 100000.0 //100km
	d := 3 * (rad - a.Location.Distance(to.Location)) / rad
	if d < 0 {
		d = 0
	}
	e += d
	e /= 9 //normatlize to 0..1
	return e
}

func (a Address) String() string {
	addr := []string{a.HouseNumber, a.Street, a.City, a.Region, a.Country, a.Postcode}
	return strings.Join(addr, ",")
}
