// Diglet is a set of geospatial tools focused around rendering large feature sets efficiently.
package mbt

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/buckhx/diglet/util"
	"github.com/codegangsta/cli"
)

var (
	CsvExt     = "csv"
	GeojsonExt = "geojson"
	exts       = []string{CsvExt, GeojsonExt}
)

var Cmd = cli.Command{
	Name:        "mbt",
	Aliases:     []string{"build"},
	Usage:       "Builds an mbtiles database from the input data source",
	Description: "Builds an mbtiles database from the given format",
	ArgsUsage:   "input_source",
	Action: func(c *cli.Context) { //TODO break out into functions
		// get kwargs & vars
		out := c.String("output")
		desc := c.String("desc")
		layer := c.String("layer-name")
		zmin := c.Int("min")
		zmax := c.Int("max")
		extent := c.Int("extent")
		upsert := c.Bool("upsert")
		force := c.Bool("force")
		// validate
		if len(c.Args()) == 0 || out == "" {
			util.Die(c, "input_source & --out required")
		} else if zmax < zmin || zmin < 0 || zmax > 23 {
			util.Die(c, "--max > --min, --min > 0 --max < 24 not satisfied")
		}
		// execute
		source, err := getSource(c)
		util.Check(err)
		if force {
			os.Remove(out)
		}
		tiles, err := InitTiles(out, upsert, desc, extent)
		util.Check(err)
		err = tiles.Build(source, layer, zmin, zmax)
		util.Check(err)
		// finalize
		file, _ := os.Open(out)
		defer file.Close()
		stat, _ := file.Stat()
		exp := float64(stat.Size()) / float64(1<<20)
		util.Info("%s was successfully caught!", out)
		util.Info("Diglet gained %f MB of EXP!", exp)
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
			Usage: "Maximum zoom level to build tiles for",
		},
		cli.IntFlag{
			Name:  "min, min-zoom",
			Value: 5,
			Usage: "Minimum zoom level to build tiles from",
		},
		cli.StringFlag{
			Name: "filter",
			Usage: "Only include fields keys in this comma delimited list.\t" +
				"EXAMPLE --filter name,date,case_number,id\t" +
				"NOTE all fields are lowercased and non-word chars replaced with '_'",
		},
		cli.StringFlag{
			Name:  "csv-lat",
			Usage: "Column containing a single longitude point",
		},
		cli.StringFlag{
			Name:  "csv-lon",
			Usage: "Column containing a single longitude point",
		},
		cli.StringFlag{
			Name: "csv-shape",
			Usage: "Column containing shape in geojson-like 'coordinates' form.\t" +
				"Does not support multi-geometries",
		},
		cli.StringFlag{
			Name:  "csv-delimiter",
			Value: ",",
		},
	},
}

func getSource(c *cli.Context) (source FeatureSource, err error) {
	path := c.Args()[0]
	var filter []string
	if len(c.String("filter")) > 0 {
		filter = strings.Split(c.String("filter"), ",")
	}
	ext := filepath.Ext(path)[1:]
	switch ext {
	case CsvExt:
		delim := c.String("csv-delimiter")
		fields := GeoFields{"lat": c.String("csv-lat"), "lon": c.String("csv-lon"), "shape": c.String("csv-shape")}
		if !fields.Validate() {
			err = util.Errorf("csv-lat/csv-lon or csv-shape required")
			break
		}
		source = NewCsvSource(path, filter, delim, fields)
	case GeojsonExt:
		source = NewGeojsonSource(path, filter)
	default:
		err = util.Errorf("Invalid source file extension %s %s", ext, strings.Join(exts, "|"))
	}
	return
}
