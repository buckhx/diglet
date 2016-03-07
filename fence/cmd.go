package fence

import (
	"bufio"
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	"github.com/codegangsta/cli"
	"os"
	"runtime"
	"strings"
	"sync"
)

var fences = []string{RtreeFence, BruteForceFence, QuadTreeFence, QuadRtreeFence}

var Cmd = cli.Command{
	Name:        "fence",
	Usage:       "Fence geojson features from stdin",
	Description: "",
	ArgsUsage:   "fence_file",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "fence",
			Value: "rtree",
			Usage: "Type of fence to use " + strings.Join(fences, "|"),
		},
		cli.IntFlag{
			Name:  "concurrency, c",
			Value: runtime.GOMAXPROCS(0),
			Usage: "Concurrency factor, defaults to number of cores",
		},
	},
	Action: func(c *cli.Context) {
		args := c.Args()
		if len(args) < 1 || args[0] == "" {
			util.Die(c, "fence_file required")
		}
		fence, err := GetFence(c.String("fence"))
		if err != nil {
			util.Die(c, err.Error())
		}
		source := geo.NewGeojsonSource(args[0], nil)
		//TODO add workers here
		features, _ := source.Publish()
		for feature := range features {
			fence.Add(feature)
		}
		queries := make(chan *geo.Feature, 1<<10)
		go func() {
			defer close(queries)
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				msg := scanner.Text()
				gj := geo.UnmarshalGeojsonFeature(msg)
				feature := geo.GeojsonFeatureAdapter(gj)
				queries <- feature
			}
		}()
		workers := c.Int("c")
		wg := &sync.WaitGroup{}
		matchs := make(chan *geo.Feature)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for query := range queries {
					for _, match := range fence.Get(query.Geometry[0].Head()) {
						matchs <- match
					}
				}
			}()
		}
		go func() {
			wg.Wait()
			close(matchs)
		}()
		for match := range matchs {
			util.Info("%+v", match.Properties["neighborhood"])
		}
	},
}
