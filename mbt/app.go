package mbt

import (
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/mbt/tile_system"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/mbtiles"
)

type Tiles struct {
	tileset *mbtiles.Tileset
	args    map[string]string
}

func InitTiles(mbtpath string, upsert bool, desc string, extent uint) (tiles Tiles, err error) {
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
	tiles = Tiles{tileset: ts}
	return
}

func (t Tiles) Build(source FeatureSource, layerName string, zmin, zmax uint) (err error) {
	for zoom := zmax; zoom >= zmin; zoom-- {
		//TODO goroutine per level
		util.Info("Generating tiles for zoom level: %d", zoom)
		features, err := source.Publish(4) //TODO cores
		if err != nil {
			return err //shadowed
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
				return err //shadowed
			}
			t.tileset.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
		}
	}
	return
}
