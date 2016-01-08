// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package main

import (
	"fmt"
	"os"

	"github.com/buckhx/diglet/mbt"
	"github.com/buckhx/diglet/resources"
	"github.com/buckhx/diglet/wms"

	"github.com/codegangsta/cli"
)

//go:generate go run scripts/include.go

func main() {
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
	app.Version = resources.Version()
	app.Commands = []cli.Command{
		{
			Name:        "start",
			Usage:       "diglet start --mbtiles path/to/tiles",
			Description: "Starts the diglet server",
			Action: func(c *cli.Context) {
				p := c.String("port")
				mbtiles := c.String("mbtiles")
				if mbtiles == "" {
					cli.ShowSubcommandHelp(c)
					die("ERROR: --mbtiles flag is required")
				}
				server := wms.MBTServer(mbtiles, p)
				server.Run()
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port",
					Value: "8080",
					Usage: "Port to bind",
				},
				cli.StringFlag{
					Name:  "mbtiles, mbt",
					Value: "",
					Usage: "REQUIRED: Path to mbtiles to serve",
				},
			},
		},
		{
			Name: "mbt",
			Action: func(c *cli.Context) {
				in := c.String("input")
				out := c.String("output")
				if in == "" || out == "" {
					die("ERROR: --in & --out required")
				}
				mbt.GeoJsonToMbtiles(in, out)
				fmt.Println("Success!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "in, input",
				},
				cli.StringFlag{
					Name: "out, output, mbtiles",
				},
			},
		},
	}
	app.Run(args)
}
