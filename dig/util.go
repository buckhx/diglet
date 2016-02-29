package dig

import (
	"bytes"
	"github.com/antzucaro/matchr"
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	"github.com/deckarep/golang-set"
	"github.com/kpawlik/geojson"
	"github.com/reiver/go-porterstemmer"
	_ "math/rand"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func workers() int {
	w := runtime.GOMAXPROCS(0)
	util.Info("Dispatching %d workers", w)
	return w
}

func reverse(ids []int64) {
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
}

func printGeojson(regions map[int64]*geo.Feature) {
	features := make([]*geojson.Feature, len(regions))
	i := 0
	for id, region := range regions {
		line := make([]geojson.Coordinate, len(region.Geometry[0].Coordinates))
		for i, c := range region.Geometry[0].Coordinates {
			line[i] = geojson.Coordinate{geojson.Coord(c.Lon), geojson.Coord(c.Lat)}
		}
		polygon := geojson.NewPolygon(geojson.MultiLine([]geojson.Coordinates{line}))
		feature := geojson.NewFeature(polygon, region.Properties, id)
		features[i] = feature
		i++
	}
	coll := geojson.NewFeatureCollection(features)
	s, err := geojson.Marshal(coll)
	util.Check(err)
	util.Info("%s", s)
}

func editDistance(s, o string) float64 {
	return matchr.JaroWinkler(s, o, true)
}

func editDist(from_st, from_hn, to_st, to_hn string) float64 {
	qterm := expand(clean(from_st))
	nterm := expand(clean(to_st))
	s := matchr.JaroWinkler(qterm, nterm, true)
	fhn, e1 := strconv.Atoi(from_hn)
	thn, e2 := strconv.Atoi(to_hn)
	h := matchr.JaroWinkler(from_hn+"  ", to_hn+"  ", true)
	if e1 == nil && e2 == nil { //use integer diff if both parsable
		d := fhn - thn
		if d < 0 {
			d = -d
		}
		h = float64(1000-d) / 1000.0
		if h < 0 {
			h = 0.0
		}
	}
	return 5*s + h
}

func mphones(value string) <-chan string {
	indexes := make(chan string)
	go func() {
		defer close(indexes)
		terms := mapset.NewSet()
		terms.Add("")
		//util.Info(value)
		value = expand(clean(value))
		//util.Info(value)
		words := strings.Split(value, " ")
		for _, word := range words {
			if _, ok := stopwords[word]; ok {
				continue
			}
			word = porterstemmer.StemString(word) //rune
			m1, m2 := matchr.DoubleMetaphone(word)
			for _, term := range terms.ToSlice() {
				terms.Remove(term)
				if len(m1) > 0 {
					terms.Add(term.(string) + m1)
				}
				if len(m2) > 0 {
					terms.Add(term.(string) + m2)
				}
			}
			//stem := stemmer.StemString(word) //rune
			//util.Info("%s -> %s", word, stem)
		}
		for term := range terms.Iter() {
			//util.Info("%s", term)
			indexes <- term.(string)
		}
		/*
			// fisher-yates shuffle
			v := terms.ToSlice()
			util.Info("%s", v)
			n := len(v)
			for i := n - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				swap(v, i, j)
				indexes <- joinStrings(v)
			}
		*/
	}()
	return indexes
}

func swap(v []interface{}, i, j int) {
	t := v[i]
	v[i] = v[j]
	v[j] = t
}

func joinStrings(vals []interface{}) string {
	var buf bytes.Buffer
	for _, v := range vals {
		buf.WriteString(v.(string))
	}
	return buf.String()
}

func expand(s string) string {
	s = util.Sprintf(" %s ", s)
	for i, k := range exkeys {
		v := exvals[i]
		s = strings.Replace(s, k, v, -1)
	}
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.Trim(s, " ")
	//return util.Sprintf(" %s ", s)
	return s
}

func clean(s string) string {
	s = strings.ToLower(s)
	s = nonword.ReplaceAllString(s, "")
	return s
}

var nonword = regexp.MustCompile("[^\\w ]")
var expansions = map[string]string{
	"0":      "zero ",
	"1":      "one ",
	"2":      "two ",
	"3":      "three ",
	"4":      "four ",
	"5":      "five ",
	"6":      "six ",
	"7":      "seven ",
	"8":      "eight ",
	"9":      "nine ",
	" n ":    "north ",
	" e ":    "east ",
	" s ":    "south ",
	" w ":    "west ",
	"north":  "north ", //for
	"east":   "east ",
	"south":  "south ",
	"west":   "west ",
	" nw ":   "north west ",
	" ne ":   "north east",
	" sw ":   "south west ",
	" se ":   "south east ",
	"first":  "one ",
	"second": "two ",
	"third":  "three ",
	"fourth": "four ",
	"fifth":  "five ",
	// more?
}

// slice iteration is much fater than map
var exkeys, exvals = expandmap(expansions)

func expandmap(expansions map[string]string) (keys, vals []string) {
	keys = make([]string, len(expansions))
	vals = make([]string, len(expansions))
	i := 0
	for k, v := range expansions {
		keys[i] = k
		vals[i] = v
		i++
	}
	return
}

var stopwords = map[string]bool{
//"nd": true,
//"th": true,
//"st": true,
}
