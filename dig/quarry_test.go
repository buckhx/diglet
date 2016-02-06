package dig_test

import (
	"github.com/buckhx/diglet/dig"
	_ "github.com/buckhx/diglet/util"
	//"sync"
	"testing"
)

func TestQuarryExcavate(t *testing.T) {
	qdb, err := dig.NewQuarry(HI_DIG)
	if err != nil {
		t.Error(err)
	}
	//qdb.Excavate(HI_PBF)
	//qdb.PrintStats()
	qdb.Dig("1000", "Bishop Street")

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
