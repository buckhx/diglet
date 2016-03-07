package fence

import (
	"bufio"
	"encoding/json"
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
		cli.IntFlag{
			Name:  "zoom, z",
			Value: 14,
			Usage: "Some fences require a zoom level",
		},
	},
	Action: func(c *cli.Context) {
		args := c.Args()
		if len(args) < 1 || args[0] == "" {
			util.Die(c, "fence_file required")
		}
		fence, err := GetFence(c.String("fence"), c.Int("z"))
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
		results := make(chan *geo.Feature)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for query := range queries {
					matchs := fence.Get(query.Geometry[0].Head())
					fences := make([]map[string]interface{}, len(matchs))
					for i, match := range matchs {
						fences[i] = match.Properties
					}
					query.Properties["fences"] = fences
					results <- query
				}
			}()
		}
		go func() {
			wg.Wait()
			close(results)
		}()
		for res := range results {
			out, err := json.Marshal(res)
			util.Check(err)
			util.Printf("%s\n", out)
			//util.Info("%+v", match.Properties["neighborhood"])
		}
	},
}
