package mbtiles

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func check(err error) {
	if err != nil {
		panic(err)
		//log.Fatal(err)
	}
}

func dbConnect(path string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", path)
	return
}

func dbReadMetadata(db *sql.DB) (md *Metadata, err error) {
	rows, err := db.Query("SELECT name, value FROM metadata")
	if err != nil {
		md = nil
		return
	}
	defer rows.Close()
	attrs := make(map[string]string)
	for rows.Next() {
		var name, value string
		rows.Scan(&name, &value)
		attrs[name] = value
	}
	md = &Metadata{attrs}
	return
}

func dbReadTile(db *sql.DB, tile *Tile) (err error) {
	stmt := "SELECT tile_data FROM tiles WHERE zoom_level=%d AND tile_column=%d AND tile_row=%d"
	q := fmt.Sprintf(stmt, tile.Z, tile.X, tile.Y)
	row := db.QueryRow(q)
	var blob []byte
	err = row.Scan(&blob)
	tile.Data = blob
	return
}

func dbWriteTile(db *sql.DB, tile *Tile) (err error) {
	stmt := "INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(stmt, tile.Z, tile.X, tile.Y, tile.Data)
	return
}

func dbInitTileset(path string, metadata *Metadata) (db *sql.DB, err error) {
	if !isPathAvailable(path) {
		return nil, fmt.Errorf("Path not available to create mbtiles: %s", path)
	}
	db, err = dbConnect(path)
	if err != nil {
		return
	}
	if !metadata.HasRequiredKeys() {
		err = fmt.Errorf("Tileset Metadata keyset is missing required keys: %v", MetadataRequiredKeys)
		return
	}
	meta_string := ""
	for k, v := range metadata.Attributes() {
		meta_string += fmt.Sprintf(" (%q, %q),", k, v)
	}
	creates := []string{
		"CREATE TABLE metadata (name text, value text)",
		"CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob)",
		"CREATE TABLE grids (zoom_level integer, tile_column integer, tile_row integer, grid blob)",
		"CREATE TABLE grid_data (zoom_level integer, tile_column integer, tile_row integer, key_name text, key_json text)",
		"INSERT INTO metadata (name, value) VALUES " + meta_string[:len(meta_string)-1],
	}
	err = dbExecStatements(db, creates...)
	return
}

func dbExecStatements(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
