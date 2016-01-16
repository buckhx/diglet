package mbt

import (
	"encoding/csv"
	"github.com/buckhx/diglet/util"
	"io"
	"os"
	"strconv"
)

type FeatureSource interface {
	Publish() (chan *Feature, error)
}

type GeoFields map[string]string

func (g GeoFields) Validate() error {
	return nil
}

type CsvSource struct {
	path      string
	headers   map[string]int
	delimiter string
	fields    GeoFields
}

func NewCsvSource(path, delimiter string, fields GeoFields) *CsvSource {
	return &CsvSource{
		path:      path,
		delimiter: delimiter,
		fields:    fields,
	}
}

func (c *CsvSource) Publish() (features chan *Feature, err error) {
	lines, err := c.publishLines()
	if err != nil {
		return
	}
	features = make(chan *Feature, 1000)
	go func() {
		defer close(features)
		for line := range lines {
			if feature, err := c.featureAdapter(line); err != nil {
				util.Warn(err, "feature adapter")
			} else {
				features <- feature
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
	defer f.Close()
	reader := csv.NewReader(f)
	c.headers = readHeaders(reader)
	err = c.fields.Validate() //c.headers)
	//TODO if err != nil
	lines = make(chan []string, 100)
	go func() {
		defer close(lines)
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.Warn(err, "line reading")
			}
			lines <- line
		}
	}()
	return
}

func (c *CsvSource) featureAdapter(line []string) (feature *Feature, err error) {
	feature = NewFeature("point")
	props := make(map[string]interface{})
	for k, i := range c.headers {
		props[k] = line[i]
	}
	feature.Properties = props
	lat, err := strconv.ParseFloat(line[c.headers[c.fields["lat"]]], 64)
	if err != nil {
		return nil, err
	}
	lon, err := strconv.ParseFloat(line[c.headers[c.fields["lon"]]], 64)
	if err != nil {
		return nil, err
	}
	point := NewShape(Coordinate{Lat: lat, Lon: lon})
	feature.AddShape(point)
	return
}

func readHeaders(reader *csv.Reader) (headers map[string]int) {
	line, err := reader.Read()
	util.Warn(err, "reading headers")
	headers = make(map[string]int, len(line))
	for i, k := range line {
		//if _, ok := c.fields[k]; !ok {
		headers[k] = i
		//}
	}
	return
}
