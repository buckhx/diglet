package tms

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/buckhx/diglet/util"
	"github.com/codegangsta/cli"
)

var Cmd = cli.Command{
	Name:        "tms",
	Aliases:     []string{"serve"},
	Usage:       "Starts the diglet Tile Map Service",
	Description: "Starts the diglet Tile Map Service. Uses Slippy maps tilenames by default.",
	ArgsUsage:   "mbtiles_directory",
	Action: func(c *cli.Context) {
		port := c.String("port")
		args := c.Args()
		if len(args) < 1 {
			util.Die(c, "directory path to serve mbtiles from is required")
		}
		mbt := args[0]
		if mbt == "" {
			util.Die(c, "mbtiles_directory is required")
		}
		cert := c.String("cert")
		key := c.String("key")
		server, err := MBTServer(mbt, port)
		if err != nil {
			util.Die(c, err.Error())
		}
		if (cert != "") && (key != "") {
			server.RunTLS(&cert, &key)
		} else if cert != "" || key != "" {
			util.Die(c, "Both cert & key are required to serve over TLS/SSL")
		} else {
			sigs := make(chan os.Signal, 1)
			done := make(chan bool, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			defer os.Remove(filepath.Join(mbt, CacheName))
			go func() {
				err := server.Run()
				util.Error(err)
				done <- true
			}()
			go func() {
				<-sigs
				done <- true
			}()
			<-done
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
		cli.BoolFlag{
			Name:  "tms-origin",
			Usage: "NOT IMPLEMENTED: Use TMS origin, SW origin w/ Y increasing North-wise. Default uses NW origin inscreasing South-wise (Slippy tilenames)",
		},
	},
}
