package mbtiles

import (
	r "reflect"
)

type Format string

const (
	PNG     Format = "PNG"
	JPG     Format = "JPG"
	GIF     Format = "GIF"
	WEBP    Format = "WEBP"
	PBF_GZ  Format = "PBF_GZ"
	PBF_DF  Format = "PBF_DF"
	UNKNOWN Format = "UNKNOWN"
	EMPTY   Format = "EMPTY"
)

// Tiles are meant to follow the TMS standard, but there are methods to Read/Write OSM style tiles in Tileset
type Tile struct {
	Z    int    `json:"z"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Data []byte `json:"data"`
}

func EmptyTile(z, x, y int) (tile *Tile) {
	return &Tile{Z: z, X: x, Y: y}
}

func (t *Tile) SniffFormat() (f Format) {
	switch {
	case len(t.Data) < 1:
		f = EMPTY
	case len(t.Data) >= 8 && r.DeepEqual(t.Data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		f = PNG
	case len(t.Data) >= 4 && r.DeepEqual(t.Data[:2], []byte{0xFF, 0xD8}) && r.DeepEqual(t.Data[len(t.Data)-2:], []byte{0xFF, 0xD9}):
		f = JPG
	case len(t.Data) >= 6 && r.DeepEqual(t.Data[:4], []byte{0x47, 0x49, 0x46, 0x38}) && (t.Data[4] == 0x39 || t.Data[4] == 0x37) && t.Data[5] == 0x61:
		f = GIF
	case len(t.Data) >= 12 && r.DeepEqual(t.Data[:4], []byte{0x52, 0x49, 0x46, 0x46}) && r.DeepEqual(t.Data[8:12], []byte{0x57, 0x45, 0x42, 0x50}):
		f = WEBP
	case len(t.Data) >= 2 && r.DeepEqual(t.Data[:2], []byte{0x78, 0x9C}):
		f = PBF_DF
	case len(t.Data) >= 2 && r.DeepEqual(t.Data[:2], []byte{0x1F, 0x8B}):
		f = PBF_GZ
	default:
		f = UNKNOWN
	}
	return
}

func (t *Tile) Equals(that *Tile) bool {
	return r.DeepEqual(t, that)
}
