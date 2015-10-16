package digletts

import (
	"log"
	"regexp"
	"strings"
)

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
