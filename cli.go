// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package main

import (
	"fmt"
	"os"

	"github.com/buckhx/diglet/mbt"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/diglet/wms"

	"github.com/codegangsta/cli"
	//"github.com/davecheney/profile"
)

//go:generate go run scripts/include.go

func main() {
	//defer profile.Start(profile.CPUProfile).Stop()
	client(os.Args)
}

func die(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

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
					cli.ShowSubcommandHelp(c)
					die("ERROR: --mbtiles flag is required")
				}
				cert := c.String("cert")
				key := c.String("key")
				server := wms.MBTServer(mbt, port)
				if (cert != "") && (key != "") {
					server.RunTLS(&cert, &key)
				} else if cert != "" || key != "" {
					cli.ShowSubcommandHelp(c)
					die("ERROR: Both cert & key are required to serve over TLS/SSL")
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
			Name: "mbt",
			Action: func(c *cli.Context) {
				in := c.String("input")
				out := c.String("output")
				extent := uint(c.Int("extent"))
				if in == "" || out == "" {
					die("ERROR: --in & --out required")
				}
				mbt.GeoJsonToMbtiles(in, out, extent)
				fmt.Println("Success!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "in, input",
				},
				cli.StringFlag{
					Name: "out, output, mbtiles",
				},
				cli.IntFlag{
					Name:  "extent",
					Value: 4096,
					Usage: "Extent of tiles to be built. Default is 4096",
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug",
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
