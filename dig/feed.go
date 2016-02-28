package dig

import (
	"encoding/csv"
	"github.com/buckhx/diglet/util"
	"io"
	"os"
	"strings"
)

func CsvFeed(path string, headers Address, delim rune) {
	f, err := os.Open(path)
	util.Check(err)
	reader := csv.NewReader(f)
	reader.Comma = delim
	queries := make(chan Address, 1<<10)
	go func() {
		defer close(queries)
		fields := headerIndexes(reader)
		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.Warn(err, util.Sprintf("%s", rec))
			}
			q := Address{}
			for i, v := range rec {
				switch i {
				case fields["addr"]:
					terms := strings.Split(v, " ")
					hn := terms[0]
					st := strings.Join(terms[1:], " ")
					q.HouseNumber = hn
					q.Street = st
				case fields["city"]:
					q.City = v
				case fields["region"]:
					q.Region = v
				case fields["postcode"]:
					q.Postcode = v
				case fields["country"]:
					q.Country = "US"
				}
			}
			//util.Info("%v", q)
			queries <- q
		}
	}()
	quarry, err := OpenQuarry("US_NY.dig")
	util.Check(err)
	matchs := quarry.DigFeed(queries)
	for match := range matchs {
		util.Info("%s", match)
	}
}

func headerIndexes(reader *csv.Reader) map[string]int {
	headers := make(map[string]int)
	rec, err := reader.Read()
	util.Info("HEADERS: %v", rec)
	util.Check(err)
	headers["addr"] = 0
	headers["city"] = 2
	headers["region"] = 3
	headers["postcode"] = 4
	headers["country"] = 5
	return headers
}
