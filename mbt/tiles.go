package mbt

import (
	"github.com/buckhx/diglet/geo/tile_system"
	"github.com/buckhx/diglet/mbt/mvt"
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

func (t Tiles) Build(source FeatureSource, layerName string, zmin, zmax int) (err error) {
	//TODO goroutine per level
	features, err := source.Publish(1) //TODO cores
	if err != nil {
		return err //shadowed
	}
	c, err := newFeatureCache(".feature.cache")
	if err != nil {
		return err //shadowed
	}
	defer c.close()
	c.indexFeatures(features)
	for tf := range c.tileFeatures(zmin, zmax) {
		tile := tf.t
		features := tf.f
		aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
		aLayer := aTile.NewLayer(layerName, tile_system.TileSize)
		for _, feature := range features {
			//fmt.Println(feature, "\n", tile, "\n")
			aFeature := MvtAdapter(feature, tile)
			aLayer.AddFeature(aFeature)
		}
		gz, err := aTile.GetTileGz()
		if err != nil {
			return err //shadowed
		}
		//fmt.Println(tile, tile.QuadKey(), "\n")
		t.tileset.WriteOSMTile(tile.IntX(), tile.IntY(), tile.IntZ(), gz)
	}
	return
}
