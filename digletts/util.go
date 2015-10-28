package digletts

import (
	"log"
	"regexp"
	"strings"

	"github.com/buckhx/mbtiles"
)

func info(format string, vals ...string) {
	log.Printf("Diglet info: "+format, vals)

}
func warn(err error, extra string) {
	if err != nil {
		log.Printf("Diglet warning: %s - %s", err, extra)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal("Diglet error: %s", err)
	}
}

func checks(errs ...error) {
	for _, err := range errs {
		check(err)
	}
}

var slugger = regexp.MustCompile("[^a-z0-9]+")

func slugged(s string) string {
	return strings.Trim(slugger.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

type header struct {
	key, value string
}

var formatEncoding = map[mbtiles.Format][]header{
	mbtiles.PNG:     []header{header{"Content-Type", "image/png"}},
	mbtiles.JPG:     []header{header{"Content-Type", "image/jpeg"}},
	mbtiles.GIF:     []header{header{"Content-Type", "image/gif"}},
	mbtiles.WEBP:    []header{header{"Content-Type", "image/webp"}},
	mbtiles.PBF_GZ:  []header{header{"Content-Type", "application/x-protobuf"}, header{"Content-Encoding", "gzip"}},
	mbtiles.PBF_DF:  []header{header{"Content-Type", "application/x-protobuf"}, header{"Content-Encoding", "deflate"}},
	mbtiles.UNKNOWN: []header{header{"Content-Type", "application/octet-stream"}},
	mbtiles.EMPTY:   []header{header{"Content-Type", "application/octet-stream"}},
}
