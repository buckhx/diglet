package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/mbtiles"
	"path/filepath"
)

type Tiles struct {
	source  FeatureSource
	tileset *mbtiles.Tileset
	args    map[string]string
}

func InitTiles(srcpath, mbtpath string, upsert bool, filter []string, desc string, extent uint) (tiles Tiles, err error) {
	tile_system.TileSize = extent
	var ts *mbtiles.Tileset
	if upsert {
		ts, err = mbtiles.ReadTileset(mbtpath)
	} else {
		attrs := map[string]string{
			"name":        util.SlugBase(mbtpath),
			"type":        "overlay",
			"version":     "1",
			"description": desc,
			"format":      "pbf.gz",
		}
		ts, err = mbtiles.InitTileset(mbtpath, attrs)
	}
	if err != nil {
		return
	}
	src := getSource(srcpath, filter)
	tiles = Tiles{
		tileset: ts,
		source:  src,
	}
	return
}

func (t Tiles) Build(layerName string, zmin, zmax uint) (err error) {
	for zoom := zmax; zoom >= zmin; zoom-- {
		//TODO goroutine per level
		util.Info("Generating tiles for zoom level: %d", zoom)
		features, err := t.source.Publish()
		if err != nil {
			return err
		}
		tiles := splitFeatures(features, zoom)
		for tile, features := range tiles {
			aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
			aLayer := aTile.NewLayer(layerName, tile_system.TileSize)
			for _, feature := range features {
				aFeature := feature.ToMvtAdapter(tile)
				aLayer.AddFeature(aFeature)
			}
			gz, err := aTile.GetTileGz()
			if err != nil {
				return err
			}
			t.tileset.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
		}
	}
	return
}

func getSource(mbtpath string, filter []string) FeatureSource {
	src := filepath.Ext(mbtpath)[1:]
	switch src {
	case "csv":
		//return NewCsvSource(mbtpath, delimiter, GeoFields{"lat": latField, "lon": lonField})
		return NewCsvSource(mbtpath, filter, ",", GeoFields{"lat": "latitude", "lon": "longitude"})
	case "geojson":
		return NewGeojsonSource(mbtpath, filter)
	default:
		return nil
	}
}

/*
func CsvTiles(path, delimiter, latField, lonField string) FeatureSource {
	return NewCsvSource(path, delimiter, GeoFields{"lat": latField, "lon": lonField})
}
func GeojsonTiles(path string) FeatureSource {
	return NewGeojsonSource(path)
}
*/
/*
func GeojsonTileset(ts *mbtiles.Tileset, gjpath string, zmin, zmax uint) {
	collection := readGeoJson(gjpath)
	for zoom := zmax; zoom >= zmin; zoom-- {
		util.Info("Generating tiles for zoom level: %d", zoom)
		tiles := splitFeatures(publishFeatureCollection(collection), zoom)
		for tile, features := range tiles {
			aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
			aLayer := aTile.NewLayer("denver", tile_system.TileSize)
			for _, feature := range features {
				aFeature := feature.ToMvtAdapter(tile)
				aLayer.AddFeature(aFeature)
			}
			gz, err := aTile.GetTileGz()
			if err != nil {
				panic(err)
			}
			ts.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
		}
	}
}
*/
