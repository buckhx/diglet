package digletts

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

func sprintSizeOf(v interface{}) string {
	return strconv.Itoa(binary.Size(v))
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func cleanTilesetName(path string) (slug string) {
	f := filepath.Base(path)
	f = strings.TrimSuffix(f, filepath.Ext(f))
	slug = slugged(f)
	return
}

func assertString(params map[string]interface{}, key string) (v string, err error) {
	v, ok := params[key].(string)
	if !ok {
		err = fmt.Errorf("Cannot parse param %q %q", key, params[key])
	}
	return
}

func assertInt(params map[string]interface{}, key string) (v int, err error) {
	fv, ok := params[key].(float64)
	if !ok {
		err = fmt.Errorf("Cannot parse param %q %q", key, params[key])
	}
	v = int(fv)
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
