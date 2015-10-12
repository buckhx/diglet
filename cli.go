package main

import (
	"os"

	"github.com/buckhx/diglet"
	"github.com/codegangsta/cli"
)

func main() {
	client(os.Args)
}

func client(args []string) {
	app := cli.NewApp()
	app.Name = "diglet"
	app.Usage = "Your friend in the tile business"
	app.Version = "dev"
	app.Commands = []cli.Command{
		{
			Name:        "start",
			Description: "Starts the diglet server",
			Action: func(c *cli.Context) {
				p := c.String("port")
				mbt := c.String("mbtiles")
				s := diglet.NewServer(mbt, p)
				s.Start()

			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port",
					Value: "8080",
					Usage: "Port to bind",
				},
				cli.StringFlag{
					Name:  "mbtiles",
					Usage: "Path to mbtiles to serve",
				},
			},
		},
	}
}
