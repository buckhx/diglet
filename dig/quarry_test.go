package dig_test

import (
	"github.com/buckhx/diglet/dig"
	_ "github.com/buckhx/diglet/util"
	//"sync"
	"testing"
)

const (
	NY_PBF  = "/vagrant/us_ny.osm.pbf"
	HI_PBF  = "/vagrant/us_hi.pbf"
	NY_ADDR = "/vagrant/ny_addresses.csv"
	NY_DIG  = "US_NY.dig"
	HI_DIG  = "US_HI.dig"
	GNPOST  = "/vagrant/postcodes/allCountries.txt"
)

func testQuarryExcavate(t *testing.T) {
	q, err := dig.OpenQuarry(NY_DIG)
	if err != nil {
		t.Error(err)
	}
	q.Excavate(NY_PBF, GNPOST)
	addr := dig.Address{HouseNumber: "72", Street: "N 4th Street", Postcode: "11249"}
	q.Dig(addr)
}

func testDig(t *testing.T) {
	q, err := dig.OpenQuarry(NY_DIG)
	if err != nil {
		t.Error(err)
	}
	addr := dig.Address{
		HouseNumber: "2154",
		Street:      "hazard hill road",
		City:        "Brooklyn",
		Region:      "New York",
		Country:     "US",
		Postcode:    "13903"}
	match := q.Dig(addr)
	t.Errorf("MATCH: %v", match)
}

func testCsvFeed(t *testing.T) {
	q, err := dig.OpenQuarry(NY_DIG)
	if err != nil {
		t.Error(err)
	}
	q.CsvFeed(NY_ADDR, "test_col", ',')
}
