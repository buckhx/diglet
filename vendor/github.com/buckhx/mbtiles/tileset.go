package mbtiles

import (
	"database/sql"
)

type Tileset struct {
	db *sql.DB
	m  *Metadata
}

func ReadTileset(path string) (ts *Tileset, err error) {
	db, err := dbConnect(path)
	if err != nil {
		return
	}
	md, err := dbReadMetadata(db)
	if err != nil {
		return
	}
	ts = &Tileset{db, md}
	return
}

// Creates a NEW, BLANK tileset at the given path
// metadata is requied to have the following keys: name, type, version, description, format
func InitTileset(path string, metadata map[string]string) (ts *Tileset, err error) {
	meta := &Metadata{attrs: metadata}
	db, err := dbInitTileset(path, meta)
	if err != nil {
		return
	}
	ts = &Tileset{db: db, m: meta}
	return
}

func (ts *Tileset) ReadTile(x, y, z int) (tile *Tile, err error) {
	tile = EmptyTile(z, x, y)
	if err = dbReadTile(ts.db, tile); err == sql.ErrNoRows {
		// if the row was empty, just keep an empty tile
		// and don't throw an error
		err = nil
	}
	return
}

func (ts *Tileset) WriteTile(x, y, z int, data []byte) (tile *Tile, err error) {
	tile = &Tile{X: x, Y: y, Z: z, Data: data}
	err = dbWriteTile(ts.db, tile)
	return
}

//Writes a tile that uses the NW origin like OSM
func (ts *Tileset) WriteOSMTile(x, y, z int, data []byte) (tile *Tile, err error) {
	y = flipY(y, z)
	tile = &Tile{X: x, Y: y, Z: z, Data: data}
	err = dbWriteTile(ts.db, tile)
	return
}

// TMS use SW origin(0,0), OSM uses Slippy names with NW origin
// See: http://gis.stackexchange.com/questions/116288/mbtiles-and-slippymap-tilenames
func (ts *Tileset) ReadOSMTile(x, y, z int) (tile *Tile, err error) {
	y = flipY(y, z)
	tile, err = ts.ReadTile(x, y, z)
	return
}

func (ts *Tileset) Metadata() *Metadata {
	return ts.m
}

func (ts *Tileset) Close() {
	ts.db.Close()
}
