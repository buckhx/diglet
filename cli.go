// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package main

import (
	"os"

	//"github.com/buckhx/diglet/dig"
	"github.com/buckhx/diglet/mbt"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/diglet/wms"

	"github.com/codegangsta/cli"
	"github.com/davecheney/profile"
)

var Version string

//go:generate go run scripts/include.go static/static.html
func client(args []string) {
	app := cli.NewApp()
	app.Name = "diglet"
	app.Usage = "Your friend in the tile business"
	app.Version = Version
	app.Commands = []cli.Command{
		wms.Cmd,
		mbt.Cmd,
		//dig.Cmd,
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Print debugging log lines",
		},
	}
	app.Run(args)
}

func main() {
	for _, arg := range os.Args {
		if arg == "--debug" {
			util.DEBUG = true
			config := &profile.Config{
				MemProfile: true,
				CPUProfile: true,
			}
			defer profile.Start(config).Stop()
		}
	}
	client(os.Args)
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
