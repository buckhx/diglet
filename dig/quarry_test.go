package dig_test

import (
	"github.com/buckhx/diglet/dig"
	_ "github.com/buckhx/diglet/util"
	//"sync"
	"testing"
)

func TestQuarryExcavate(t *testing.T) {
	qdb, err := dig.NewQuarry(NY_DIG)
	if err != nil {
		t.Error(err)
	}
	qdb.Excavate(NY_PBF)
	//qdb.PrintStats()
	//addr := dig.Address{HouseNumber: "72", Street: "N 4th Street", Postcode: "11249"}
	//qdb.Dig(addr)
	//qdb.Dig("11", "west 42nd Street", "")
}
