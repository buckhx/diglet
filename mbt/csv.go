package mbt

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"

	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	"github.com/deckarep/golang-set"
)

type FeatureSource interface {
	Publish() (chan *geo.Feature, error)
}

type GeoFields map[string]string

func (g GeoFields) Validate() bool {
	return g.HasCoordinates() != g.HasShape() //xor
}

func (g GeoFields) HasCoordinates() bool {
	return g["lat"] != "" && g["lon"] != ""
}

func (g GeoFields) HasShape() bool {
	return g["shape"] != ""
}

type CsvSource struct {
	path      string
	headers   map[string]int
	delimiter string
	filter    mapset.Set
	fields    GeoFields
}

func NewCsvSource(path string, filter []string, delimiter string, fields GeoFields) *CsvSource {
	var set mapset.Set
	if filter == nil || len(filter) == 0 {
		set = nil
	} else {
		set = mapset.NewSet()
		for _, k := range filter {
			set.Add(k)
		}
		set.Add(fields["lat"])
		set.Add(fields["lon"])
	}
	return &CsvSource{
		path:      path,
		delimiter: delimiter,
		filter:    set,
		fields:    fields,
	}
}

func (c *CsvSource) Publish() (features chan *geo.Feature, err error) {
	lines, err := c.publishLines()
	if err != nil {
		return
	}
	features = make(chan *geo.Feature)
	go func() {
		defer close(features)
		//TODO read ID from csv
		id := uint64(0)
		for line := range lines {
			if feature, err := c.featureAdapter(line); err != nil {
				util.Warn(err, "feature adapter")
			} else {
				if feature.ID != nil {
					feature.Properties["fid"] = feature.ID
				}
				feature.ID = id
				features <- feature
				id++
			}
		}
	}()
	return
}

func (c *CsvSource) publishLines() (lines chan []string, err error) {
	//TODO optionally trim lines
	f, err := os.Open(c.path)
	if err != nil {
		return
	}
	reader := csv.NewReader(f)
	c.headers = readHeaders(reader, c.filter)
	//TODO if err != nil
	lines = make(chan []string, 100)
	go func() {
		defer close(lines)
		defer f.Close()
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.Warn(err, "line reading")
			} else if line[c.headers[c.fields["lat"]]] == "" || line[c.headers[c.fields["geometry"]]] == "" {
				continue
				//err = util.Errorf("No coordinates %v", line)
				//util.Warn(err, "no lat/lon")
			} else {
				lines <- line
			}
		}
	}()
	return
}

func (c *CsvSource) featureAdapter(line []string) (feature *geo.Feature, err error) {
	props := make(map[string]interface{}, len(c.headers)) //biggest malloc
	for k, i := range c.headers {
		props[k] = line[i]
	}
	switch {
	case c.fields.HasCoordinates():
		feature = geo.NewPointFeature()
		feature.Properties = props
		lat, err := strconv.ParseFloat(line[c.headers[c.fields["lat"]]], 64)
		if err != nil {
			return nil, err
		}
		lon, err := strconv.ParseFloat(line[c.headers[c.fields["lon"]]], 64)
		if err != nil {
			return nil, err
		}
		point := geo.NewShape(geo.Coordinate{Lat: lat, Lon: lon})
		feature.AddShape(point)
	case c.fields.HasShape():
		g := line[c.headers[c.fields["shape"]]]
		shp, err := geo.ShapeFromString(g)
		if err != nil {
			return nil, util.Errorf("Invalid shape format %+v", g)
		}
		switch {
		case len(shp.Coordinates) == 0:
			feature = geo.NewPointFeature()
		case len(shp.Coordinates) == 1:
			feature = geo.NewPointFeature()
		case shp.Coordinates[0] == shp.Coordinates[len(shp.Coordinates)-1]: //closed
			feature = geo.NewPolygonFeature()
		default:
			feature = geo.NewLineFeature()
		}
		feature.Properties = props
		feature.AddShape(shp)
	default:
		err = util.Errorf("Invalid line")
	}
	return
}

func readHeaders(reader *csv.Reader, filter mapset.Set) (headers map[string]int) {
	line, err := reader.Read()
	util.Warn(err, "reading headers")
	headers = make(map[string]int, len(line))
	for i, k := range line {
		//if _, ok := c.fields[k]; !ok {
		k = util.Slugged(k, "_")
		if filter == nil || filter.Contains(k) {
			headers[k] = i
		}
		//}
	}
	util.Debug("Headers %v", headers)
	return
}
