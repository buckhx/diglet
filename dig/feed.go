package dig

import (
	"encoding/csv"
	"github.com/buckhx/diglet/util"
	"io"
	"os"
)

func csvFeed(path, col string, delim rune) <-chan Address {
	f, err := os.Open(path)
	util.Check(err)
	reader := csv.NewReader(f)
	reader.Comma = delim
	queries := make(chan Address, 1<<10)
	go func() {
		defer close(queries)
		idx := colidx(reader, col)
		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				util.Warn(err, util.Sprintf("%s", rec))
			}
			queries <- StringAddress(rec[idx])
			/*
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
			*/
		}
	}()
	return queries
}

func colidx(reader *csv.Reader, col string) int {
	rec, err := reader.Read()
	util.Check(err)
	for i, k := range rec {
		if k == col {
			return i
		}
	}
	util.Fatal("No clum named %q in headers %v", col, rec)
	return -1
	/*
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
	*/
}
