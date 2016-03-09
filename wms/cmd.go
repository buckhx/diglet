package wms

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/buckhx/diglet/util"
	"github.com/codegangsta/cli"
)

var Cmd = cli.Command{
	Name:        "wms",
	Aliases:     []string{"serve"},
	Usage:       "Starts the diglet Web Map Service",
	Description: "Starts the diglet Web Map Service",
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
		server := MBTServer(mbt, port)
		if (cert != "") && (key != "") {
			server.RunTLS(&cert, &key)
		} else if cert != "" || key != "" {
			util.Die(c, "Both cert & key are required to serve over TLS/SSL")
		} else {
			sigs := make(chan os.Signal, 1)
			done := make(chan bool, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			defer os.Remove(mbt + "/" + CacheName) //TODO make path.Join
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
			Usage: "NOT IMPLEMENTED: Use TMS origin, SW origin w/ Y increasing North-wise",
		},
	},
}
