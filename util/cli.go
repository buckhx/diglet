package util

import (
	"github.com/codegangsta/cli"
)

func Die(c *cli.Context, msg string) {
	cli.ShowSubcommandHelp(c)
	Fatal(msg)
}
