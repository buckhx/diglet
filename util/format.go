package util

import (
	"path/filepath"
	"regexp"
	"strings"
)

var slugger = regexp.MustCompile("[^a-z0-9]+")

func Slug(s string) string {
	return Slugged(s, "-")
}

func Slugged(s, delim string) string {
	return strings.Trim(slugger.ReplaceAllString(strings.ToLower(s), delim), delim)
}

// Slugs the basename of the path, removing the path and extension
// "/path/to/file_2.gz " -> "file-2"
func SlugBase(path string) (slug string) {
	f := filepath.Base(path)
	f = strings.TrimSuffix(f, filepath.Ext(f))
	return Slug(f)
}
