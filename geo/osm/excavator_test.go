package osm_test

import (
	"github.com/buckhx/diglet/geo/osm"
	"github.com/buckhx/diglet/util"
	//"sync"
	_ "testing"
)

var (
	NY_PBF = "/vagrant/us_ny.osm.pbf"
	HI_PBF = "/vagrant/us_hi.pbf"
	NY_DIG = "US_NY.dig"
	HI_DIG = "US_HI.dig"
	GNPOST = "/vagrant/postcodes/allCountries.txt"
)

func noder(nodes <-chan *osm.Node) {
	for node := range nodes {
		util.Info("Node: %d %s", node.ID, node.Tags)
	}
}

func wayer(ways <-chan *osm.Way) {
	for way := range ways {
		util.Info("Way: %d %s", way.ID, way.Tags)
	}
}

/*
func testExcavate(t *testing.T) {
	ex, err := dig.NewExcavator(HI_PBF)
	//ex.NodeFunc = noder
	//ex.WayFunc = wayer
	if err != nil {
		t.Error(err)
	}
	err = ex.Start(1)
	if err != nil {
		t.Error(err)
	}
}
*/
