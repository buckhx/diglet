package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/mbtiles"
)

func CreateTileset(mbtpath, desc string, extent uint) (ts *mbtiles.Tileset, err error) {
	tile_system.TileSize = extent
	attrs := map[string]string{
		"name":        util.SlugBase(mbtpath),
		"type":        "overlay",
		"version":     "1",
		"description": desc,
		"format":      "pbf.gz",
	}
	ts, err = mbtiles.InitTileset(mbtpath, attrs)
	return
}

func CsvTiles(path, delimiter, latField, lonField string) FeatureSource {
	return NewCsvSource(path, delimiter, GeoFields{"lat": latField, "lon": lonField})
}
func GeojsonTiles(path string) FeatureSource {
	return NewGeojsonSource(path)
}

func BuildTileset(ts *mbtiles.Tileset, source FeatureSource, zmin, zmax uint) {
	for zoom := zmax; zoom >= zmin; zoom-- {
		//TODO goroutine per level
		util.Info("Generating tiles for zoom level: %d", zoom)
		features, err := source.Publish()
		util.Check(err)
		tiles := splitFeatures(features, zoom)
		for tile, features := range tiles {
			aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
			aLayer := aTile.NewLayer("features", tile_system.TileSize)
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
