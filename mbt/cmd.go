// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package mbt

import (
	"os"
	"strings"

	"github.com/buckhx/diglet/util"

	"github.com/codegangsta/cli"
)

var Cmd = cli.Command{
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
			util.Die(c, "an input data source is required")
		}
		in := args[0]
		if in == "" || out == "" {
			util.Die(c, "--in & --out required")
		}
		if zmax < zmin || zmin < 0 || zmax > 23 {
			util.Die(c, "--max > --min, --min > 9 --max < 24 not satisfied")
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
		var filter []string
		if len(c.String("filter")) > 0 {
			filter = strings.Split(c.String("filter"), ",")
		}
		if tiles, err := InitTiles(in, out, upsert, filter, desc, extent); err != nil {
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
			Name:  "csv-geometry",
			Value: "geometry",
			Usage: "Column containing geometry in geojson-like 'coordinates' form",
		},
		cli.StringFlag{
			Name:  "csv-delimiter",
			Value: ",",
		},
	},
}
