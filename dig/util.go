package dig

import (
	"bytes"
	"github.com/antzucaro/matchr"
	_ "github.com/buckhx/diglet/util"
	"github.com/deckarep/golang-set"
	"github.com/reiver/go-porterstemmer"
	_ "math/rand"
	"regexp"
	"strings"
)

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
	for i, o := range expansions {
		s = strings.Replace(s, i, o, -1)
	}
	s = strings.Replace(s, "  ", " ", -1)
	//return util.Sprintf(" %s ", s)
	return s
}

func clean(s string) string {
	s = strings.ToLower(s)
	return nonword.ReplaceAllString(s, "")
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

var stopwords = map[string]bool{
//"nd": true,
//"th": true,
//"st": true,
}
