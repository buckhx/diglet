// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package dig

import (
	"github.com/buckhx/diglet/util"
	"github.com/codegangsta/cli"
)

var Cmd = cli.Command{
	Name:        "dig",
	Aliases:     []string{"geocode"},
	Usage:       "Geocoding utility",
	Description: "Geocoding",
	ArgsUsage:   "digdb",
	Action: func(c *cli.Context) {
		args := c.Args()
		if len(args) < 1 {
			util.Die(c, "requires a digdb (*.dig)")
		}
		defer util.Info("Done!")
		db := args[0]
		pbf := c.String("pbf")
		gn := c.String("geonames")
		csv := c.String("csv")
		d := c.String("csv-delimiter")
		q := c.String("query")
		quarry, _ := OpenQuarry(db)
		if pbf != "" {
			err := quarry.Excavate(pbf, gn)
			util.Check(err)
		} else if csv != "" {
			quarry.CsvFeed(csv, q, rune(d[0]))
		} else if q != "" {
			addr := QueryAddress(q)
			match := quarry.Dig(addr)
			util.Printf("Found! %s", match)
		}
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "pbf",
			Usage: "Translate this osm pbf into the db",
		},
		cli.StringFlag{
			Name:  "geonames, gn",
			Usage: "Needed for enriching pbf with postcode",
		},
		cli.StringFlag{
			Name:  "query, q",
			Usage: "Address to geocode, if db is being served, this will block",
		},
		cli.StringFlag{
			Name:  "csv",
			Usage: "Path to csv to geocode. Use with -q to select Address column.",
		},
		cli.StringFlag{
			Name:  "csv-delimiter",
			Value: ",",
		},
	},
}
