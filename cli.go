// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package main

import (
	"os"
	"strings"

	"github.com/buckhx/diglet/dig"
	"github.com/buckhx/diglet/mbt"
	"github.com/buckhx/diglet/resources"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/diglet/wms"

	"github.com/codegangsta/cli"
	"github.com/davecheney/profile"
)

//go:generate go run scripts/include.go static/static.html
func client(args []string) {
	app := cli.NewApp()
	app.Name = "diglet"
	app.Usage = "Your friend in the tile business"
	app.Version = resources.Version()
	app.Commands = []cli.Command{
		{
			Name:        "wms",
			Aliases:     []string{"serve"},
			Usage:       "Starts the diglet Web Map Service",
			Description: "Starts the diglet server",
			ArgsUsage:   "mbtiles_directory",
			Action: func(c *cli.Context) {
				port := c.String("port")
				args := c.Args()
				if len(args) < 1 {
					die(c, "directory path to serve mbtiles from is required")
				}
				mbt := args[0]
				if mbt == "" {
					die(c, "mbtiles_directory is required")
				}
				cert := c.String("cert")
				key := c.String("key")
				server := wms.MBTServer(mbt, port)
				if (cert != "") && (key != "") {
					server.RunTLS(&cert, &key)
				} else if cert != "" || key != "" {
					die(c, "Both cert & key are required to serve over TLS/SSL")
				} else {
					server.Run()
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port",
					Value: "8080",
					Usage: "Port to bind",
				},
				cli.StringFlag{
					Name:  "cert, tls-certificate",
					Usage: "Path to .pem TLS Certificate. Both cert & key required to serve HTTPS",
				},
				cli.StringFlag{
					Name:  "key, tls-private-key",
					Usage: "Path to .pem TLS Private Key. Both cert & key required to serve HTTPS",
				},
			},
		},
		{
			Name:        "mbt",
			Aliases:     []string{"build"},
			Usage:       "Builds an mbtiles database from the input data source",
			Description: "Builds an mbtiles database from the given format",
			ArgsUsage:   "input_source",
			Action: func(c *cli.Context) {
				out := c.String("output")
				desc := c.String("desc")
				layer := c.String("layer-name")
				zmin := uint(c.Int("min"))
				zmax := uint(c.Int("max"))
				extent := uint(c.Int("extent"))
				args := c.Args()
				if len(args) < 1 {
					die(c, "an input data source is required")
				}
				in := args[0]
				if in == "" || out == "" {
					die(c, "--in & --out required")
				}
				if zmax < zmin || zmin < 0 || zmax > 23 {
					die(c, "--max > --min, --min > 9 --max < 24 not satisfied")
				}
				force := c.Bool("force")
				if force {
					os.Remove(out)
				}
				/*
						lat := c.String("csv-lat")
						lon := c.String("csv-lon")
						delim := c.String("csv-delimiter")
						source := mbt.CsvTiles(in, delim, lat, lon)
					source := mbt.GeojsonTiles(in)
					mbt.BuildTileset(ts, source, zmin, zmax)
				*/
				upsert := c.Bool("upsert")
				filter := strings.Split(c.String("filter"), ",")
				if tiles, err := mbt.InitTiles(in, out, upsert, filter, desc, extent); err != nil {
					util.Fatal(err.Error())
				} else {
					err = tiles.Build(layer, zmin, zmax)
					if err != nil {
						util.Fatal(err.Error())
					} else {
						file, _ := os.Open(out)
						defer file.Close()
						stat, _ := file.Stat()
						exp := float64(stat.Size()) / float64(1<<20)
						util.Info("%s was successfully caught!", out)
						util.Info("Diglet gained %f MB of EXP!", exp)
					}
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "o, output",
					Usage: "REQUIRED: Path to write mbtiles to",
				},
				cli.StringFlag{
					Name:  "input-type",
					Value: "sniff",
					Usage: "Type of input files, 'sniff' will pick type based on the extension",
				},
				cli.BoolFlag{
					Name:  "f, force",
					Usage: "Remove the existing .mbtiles file before running.",
				},
				cli.BoolFlag{
					Name:  "u, upsert",
					Usage: "Upsert into mbtiles instead of replacing.",
				},
				cli.StringFlag{
					Name:  "layer-name",
					Value: "features",
					Usage: "Name of the layer for the features to be added to",
				},
				cli.StringFlag{
					Name:  "desc, description",
					Value: "Generated from Diglet",
					Usage: "Value inserted into the description entry of the mbtiles",
				},
				cli.IntFlag{
					Name:  "extent",
					Value: 4096,
					Usage: "Extent of tiles to be built. Default is 4096",
				},
				cli.IntFlag{
					Name:  "max, max-zoom",
					Value: 10,
					Usage: "Maximum zoom level to build tiles for. Not Implemented.",
				},
				cli.IntFlag{
					Name:  "min, min-zoom",
					Value: 5,
					Usage: "Minimum zoom level to build tiles from. Not Implemented.",
				},
				cli.StringFlag{
					Name: "filter",
					Usage: "Only include fields keys in this comma delimited list.\t" +
						"EXAMPLE --filter name,date,case_number,id\t" +
						"NOTE all fields are lowercased and non-word chars replaced with '_'",
				},
				cli.StringFlag{
					Name:  "csv-lat",
					Value: "latitude",
				},
				cli.StringFlag{
					Name:  "csv-lon",
					Value: "longitude",
				},
				cli.StringFlag{
					Name:  "csv-delimiter",
					Value: ",",
				},
			},
		},
		{
			Name:        "dig",
			Aliases:     []string{"geocode"},
			Usage:       "Geocoding utility",
			Description: "Geocoding",
			ArgsUsage:   "digdb",
			Action: func(c *cli.Context) {
				args := c.Args()
				if len(args) < 1 {
					die(c, "requires a digdb (*.dig)")
				}
				defer util.Info("Done!")
				db := args[0]
				pbf := c.String("pbf")
				gn := c.String("geonames")
				csv := c.String("csv")
				d := c.String("csv-delimiter")
				q := c.String("query")
				quarry, _ := dig.OpenQuarry(db)
				if pbf != "" {
					err := quarry.Excavate(pbf, gn)
					util.Check(err)
				} else if csv != "" {
					quarry.CsvFeed(csv, q, rune(d[0]))
				} else if q != "" {
					addr := dig.QueryAddress(q)
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
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Print debugging log lines",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			util.DEBUG = true
		}
		return nil
	}
	app.Run(args)
}

func main() {
	for _, arg := range os.Args {
		if arg == "--debug" {
			config := &profile.Config{
				MemProfile: true,
				CPUProfile: true,
			}
			defer profile.Start(config).Stop()
		}
	}
	client(os.Args)
}

func die(c *cli.Context, msg string) {
	cli.ShowSubcommandHelp(c)
	util.Fatal(msg)
}

/*
Go! Diglet!
Diglet used Earthquake
Foe DEWGONG fainted
Diglet gain 1960 EXP. Points
ELITE FOUR LORELEI is about to use CLOYSTER
sent out
Diglet fainted!
*/
