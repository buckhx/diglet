package dig_test

import (
	"github.com/buckhx/diglet/dig"
	_ "github.com/buckhx/diglet/util"
	//"sync"
	"testing"
)

const (
	NY_PBF = "/vagrant/us_ny.osm.pbf"
	HI_PBF = "/vagrant/us_hi.pbf"
	NY_DIG = "US_NY.dig"
	HI_DIG = "US_HI.dig"
	GNPOST = "/vagrant/postcodes/allCountries.txt"
)

func TestQuarryExcavate(t *testing.T) {
	qdb, err := dig.NewQuarry(NY_DIG)
	if err != nil {
		t.Error(err)
	}
	//qdb.Survey(GNPOST)
	dig.Excavate(qdb, NY_PBF, GNPOST)

	//qdb.Excavate(NY_PBF)
	//qdb.PrintStats()
	//addr := dig.Address{HouseNumber: "72", Street: "N 4th Street", Postcode: "11249"}
	//qdb.Dig(addr)
	//qdb.Dig("11", "west 42nd Street", "")
}
