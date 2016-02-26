package dig

import (
	"encoding/csv"
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
	"os"
	"strconv"
	"strings"
)

// Postcode data from GeoNames
// http://download.geonames.org/export/zip/
type Postcode struct {
	CountryCode   string
	PostCode      string
	PlaceName     string
	RegionName    string
	RegionCode    string
	SubregionName string
	SubregionCode string
	TownName      string
	TownCode      string
	Center        geo.Coordinate
	Accuracy      uint8
	RelationKey   string
}

func (p *Postcode) Key() string {
	return strings.Join([]string{p.CountryCode, p.PostCode}, ":")
}

/*
func (p *Postcode) String() string {
	return util.Sprintf("%v", p)
}
*/

func (p *Postcode) Keyed() (k, v []byte) {
	k = []byte(p.Key())
	v, err := msgpack.Marshal(p)
	if err != nil {
		k = nil
	}
	return
}

func unmarshalPostcode(b []byte) (p *Postcode, err error) {
	err = msgpack.Unmarshal(b, &p)
	return
}

// Generate postcodes from a geonames tab-delimited csv
// http://download.geonames.org/export/zip/
func ReadPostcodes(path string) <-chan *Postcode {
	fh, err := os.Open(path)
	util.Check(err)
	reader := csv.NewReader(fh)
	reader.Comma = '\t'
	postcodes := make(chan *Postcode, 1<<10)
	go func() {
		defer close(postcodes)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.Warn(err, util.Sprintf("%s", record))
			}
			if len(record) != 12 {
				util.Warn(util.Errorf("Invalid record len: %d - %s", len(record), record), "read postcodes")
				continue
			}
			pc := &Postcode{}
			c := geo.Coordinate{}
			for i, v := range record {
				switch i {
				case 0:
					pc.CountryCode = v
				case 1:
					pc.PostCode = v
				case 2:
					pc.PlaceName = v
				case 3:
					pc.RegionName = v
				case 4:
					pc.RegionCode = v
				case 5:
					pc.SubregionName = v
				case 6:
					pc.SubregionCode = v
				case 7:
					pc.TownName = v
				case 8:
					pc.TownCode = v
				case 9:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						c.Lat = f
					}
				case 10:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						c.Lon = f
					}
				case 11:
					if long, err := strconv.ParseUint(v, 10, 8); err == nil {
						short := uint8(long)
						pc.Accuracy = short
					}
				}
			}
			pc.Center = c
			postcodes <- pc
		}
	}()
	return postcodes
}
