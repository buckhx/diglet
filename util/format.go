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

// CastUInt64 casts an int64-like interface to int64 and ok == false if not castable
func CastUInt64(n interface{}) (v uint64, ok bool) {
	ok = true
	switch n := n.(type) {
	case int:
		v = uint64(n)
	case int8:
		v = uint64(n)
	case int16:
		v = uint64(n)
	case int32:
		v = uint64(n)
	case int64:
		v = uint64(n)
	case uint:
		v = uint64(n)
	case uintptr:
		v = uint64(n)
	case uint8:
		v = uint64(n)
	case uint16:
		v = uint64(n)
	case uint32:
		v = uint64(n)
	case uint64:
		v = uint64(n)
	default:
		ok = false
	}
	return
}
