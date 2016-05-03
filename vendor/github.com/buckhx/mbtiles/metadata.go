package mbtiles

import (
	"strings"
)

const (
	NAME = "name"
	TYPE = "type"
	VERS = "version"
	DESC = "description"
	FRMT = "format"
	BNDS = "bounds"
	ATTR = "attribution"
)

var MetadataRequiredKeys = []string{NAME, TYPE, VERS, DESC, FRMT}

type Metadata struct {
	attrs map[string]string
}

// The map of the kv pairs
func (m *Metadata) Attributes() map[string]string {
	return m.attrs
}

func (m *Metadata) Name() string {
	return m.attrs[NAME]
}

func (m *Metadata) Type() string {
	return m.attrs[TYPE]
}

func (m *Metadata) Version() string {
	return m.attrs[VERS]
}

func (m *Metadata) Description() string {
	return m.attrs[DESC]
}

func (m *Metadata) Format() string {
	return m.attrs[FRMT]
}

func (m *Metadata) Bounds() ([4]Coordinate, error) {
	var bnds [4]Coordinate
	for i, coordstr := range strings.Split(m.attrs[BNDS], ",") {
		c, err := ParseCoordinate(coordstr)
		if err != nil {
			return [4]Coordinate{}, err
		}
		bnds[i] = c
	}
	return bnds, nil
}

func (m *Metadata) Attribution() string {
	return m.attrs[ATTR]
}

func (m *Metadata) HasRequiredKeys() bool {
	for _, req := range MetadataRequiredKeys {
		found := false
		for key, _ := range m.attrs {
			if req == key {
				found = true
			}
		}
		if found == false {
			return false
		}
	}
	return true
}
