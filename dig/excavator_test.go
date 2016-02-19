package dig_test

import (
	"github.com/buckhx/diglet/dig"
	"github.com/buckhx/diglet/util"
	//"sync"
	"testing"
)

var (
	NY_PBF = "/vagrant/us_ny.osm.pbf"
	HI_PBF = "/vagrant/us_hi.pbf"
	NY_DIG = "US_NY.dig"
	HI_DIG = "US_HI.dig"
	GNPOST = "/vagrant/postcodes/allCountries.txt"
)

func noder(nodes <-chan *dig.Node) {
	for node := range nodes {
		util.Info("Node: %d %s", node.ID, node.Tags)
	}
}

func wayer(ways <-chan *dig.Way) {
	for way := range ways {
		util.Info("Way: %d %s", way.ID, way.Tags)
	}
}

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
	//qdb, err := dig.NewQuarry(HI_DIG)
	//if err != nil {
	//	t.Error(err)
	//}

	/*
		util.Info("Starting couriers")
		ex.NodeWorkers(func(nodes <-chan *dig.Node) {
			for node := range nodes {
				util.Info("Node: %d %s", node.ID, node.Tags)
			}
		}, 4).Wait()
			couriers := &sync.WaitGroup{}
			couriers.Add(3)
			go func() {
				defer couriers.Done()
			}()
			go func() {
				defer couriers.Done()
				for way := range ex.Ways() {
					util.Info("Way: %d %s", way.ID, way.Tags)
				}
			}()
			go func() {
				defer couriers.Done()
				for rel := range ex.Relations() {
					util.Info("Relation: %d %s", rel.ID, rel.Tags)
				}
			}()
			go func() {
				for err := range ex.Errors() {
					util.Warn(err, "couriers")
				}
			}()
			util.Info("Waiting on couriers")
			couriers.Wait()
	*/
}
