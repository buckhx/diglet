package tms

import (
	"encoding/binary"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/buckhx/mbtiles"
)

func info(format string, vals ...interface{}) {
	log.Printf("Diglet info: "+format, vals...)

}

func warn(err error, extra string) {
	if err != nil {
		log.Printf("Diglet warning: %s - %s", err, extra)
	}
}

func errorlog(err ...error) {
	log.Printf("Diglet error: %s", err)
}

func check(err error) {
	if err != nil {
		panic(err)
		log.Fatal("Fatal Diglet error: %s", err)
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
	mbtiles.PNG:     {{"Content-Type", "image/png"}},
	mbtiles.JPG:     {{"Content-Type", "image/jpeg"}},
	mbtiles.GIF:     {{"Content-Type", "image/gif"}},
	mbtiles.WEBP:    {{"Content-Type", "image/webp"}},
	mbtiles.PBF_GZ:  {{"Content-Type", "application/x-protobuf"}, {"Content-Encoding", "gzip"}},
	mbtiles.PBF_DF:  {{"Content-Type", "application/x-protobuf"}, {"Content-Encoding", "deflate"}},
	mbtiles.UNKNOWN: {{"Content-Type", "application/octet-stream"}},
	mbtiles.EMPTY:   {{"Content-Type", "application/octet-stream"}},
}

func sprintSizeOf(v interface{}) string {
	return strconv.Itoa(binary.Size(v))
}

func sprintf(format string, vals ...interface{}) string {
	return fmt.Sprintf(format, vals...)
}

func errorf(format string, vals ...interface{}) error {
	return fmt.Errorf(format, vals...)
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func atoi(v string) (int, error) {
	return strconv.Atoi(v)
}

func atof(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

func cleanTilesetName(path string) (slug string) {
	f := filepath.Base(path)
	f = strings.TrimSuffix(f, filepath.Ext(f))
	slug = slugged(f)
	return
}

func toLower(v string) string {
	return strings.ToLower(v)
}

func assertString(v interface{}) (err error) {
	if _, ok := v.(string); !ok {
		err = fmt.Errorf("Cannot assert string %q", v)
	}
	return
}

func assertNumber(v interface{}) (err error) {
	if _, ok := v.(float64); !ok {
		err = fmt.Errorf("Cannot assert number %q", v)
	}
	return
}

func castTile(t interface{}) (tile *mbtiles.Tile, err error) {
	tile, ok := t.(*mbtiles.Tile)
	if !ok {
		err = fmt.Errorf("Cannot cast tile %q", t)
	}
	return
}

func validateParams(params map[string]interface{}, keys []string) (err error) {
	for _, key := range keys {
		if _, ok := params[key]; !ok {
			err = fmt.Errorf("Missing param: %q", key)
		}
	}
	return
}
