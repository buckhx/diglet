// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package main

import (
	"os"

	"github.com/buckhx/diglet/mbt"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/diglet/wms"

	"github.com/codegangsta/cli"
	//"github.com/davecheney/profile"
)

//go:generate go run scripts/include.go
func client(args []string) {
	app := cli.NewApp()
	app.Name = "diglet"
	app.Usage = "Your friend in the tile business"
	app.Version = util.Version()
	app.Commands = []cli.Command{
		{
			Name:        "start",
			Usage:       "diglet start --mbtiles path/to/tiles",
			Description: "Starts the diglet server",
			Action: func(c *cli.Context) {
				port := c.String("port")
				mbt := c.String("mbtiles")
				if mbt == "" {
					die(c, "--mbtiles flag is required")
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
					Name:  "mbtiles",
					Usage: "REQUIRED: Path to mbtiles to serve",
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
			Usage:       "diglet mbt --in file.geojson --out tileset.mbtiles",
			Description: "Builds an mbtiles database from the given format",
			Action: func(c *cli.Context) {
				in := c.String("input")
				out := c.String("output")
				desc := c.String("desc")
				zmin := uint(c.Int("min"))
				zmax := uint(c.Int("max"))
				extent := uint(c.Int("extent"))
				if in == "" || out == "" {
					die(c, "--in & --out required")
				}
				if zmax < zmin || zmin < 0 || zmax > 23 {
					die(c, "--max > --min, --min > 9 --max < 24 not satisfied")
				}
				/*
						lat := c.String("csv-lat")
						lon := c.String("csv-lon")
						delim := c.String("csv-delimiter")
						source := mbt.CsvTiles(in, delim, lat, lon)
					source := mbt.GeojsonTiles(in)
					mbt.BuildTileset(ts, source, zmin, zmax)
				*/
				if tiles, err := mbt.InitTiles(in, out, desc, extent); err != nil {
					util.Fatal(err.Error())
				} else {
					err = tiles.Build(zmin, zmax)
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
					Name:  "in, input",
					Usage: "REQUIRED: Path to read from",
				},
				cli.StringFlag{
					Name:  "out, output, mbtiles",
					Usage: "REQUIRED: Path to write mbtiles to",
				},
				cli.StringFlag{
					Name:  "input-type",
					Value: "sniff",
					Usage: "Type of input files, 'sniff' will pick type based on the extension",
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
					Name:  "csv-lat",
					Value: "Latitude",
				},
				cli.StringFlag{
					Name:  "csv-lon",
					Value: "Longitude",
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
	//defer profile.Start(profile.CPUProfile).Stop()
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
