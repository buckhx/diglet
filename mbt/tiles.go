package mbt

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/mbt/mvt"
	"github.com/buckhx/diglet/util"
	"github.com/buckhx/mbtiles"
	"github.com/buckhx/tiles"
)

// ClipBuffer is the number of pixels to buffer a tile clipping
var ClipBuffer = 10

type Tileset struct {
	tileset *mbtiles.Tileset
	args    map[string]string
}

func InitTiles(mbtpath string, upsert bool, desc string, extent int) (t Tileset, err error) {
	tiles.TileSize = extent
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
	t = Tileset{tileset: ts}
	return
}

func (t Tileset) Build(source FeatureSource, layerName string, zmin, zmax int) (err error) {
	features, err := source.Publish()
	if err != nil {
		return
	}
	c := newFeatureIndex()
	c.indexFeatures(features, zmax)
	vts := make(chan *mvt.TileAdapter, 1<<10)
	_ = util.NWork(func() {
		defer close(vts)
		for tf := range c.tileFeatures(zmin, zmax) {
			vt := buildTile(layerName, tf)
			vts <- vt
			if err != nil {
				return
			}
		}
	}, 1)
	wg := util.Work(func() {
		for vt := range vts {
			err := writeTile(t.tileset, vt)
			if err != nil {
				return
			}
		}
	})
	wg.Wait()
	return
}

func buildTile(layer string, tf tileFeatures) *mvt.TileAdapter {
	tile, features := tf.tileFeatures()
	util.Info("Building tile %v with %d features", tile, len(features))
	aTile := mvt.NewTileAdapter(tile.X, tile.Y, tile.Z)
	aLayer := aTile.NewLayer(layer, tiles.TileSize)
	for _, feature := range features {
		aFeature := MvtAdapter(feature, tile)
		if aFeature.Valid() {
			aLayer.AddFeature(aFeature)
		}
	}
	return aTile
}

func writeTile(tileset *mbtiles.Tileset, vt *mvt.TileAdapter) error {
	gz, err := vt.GetTileGz()
	if err != nil {
		return err
	}
	tileset.WriteOSMTile(vt.X, vt.Y, vt.Z, gz)
	return nil
}

// FeatureTiles returns a list of tiles that cover the feature at the given zoom level
// Dups are not checked for, so they can exist
func FeatureTiles(f *geo.Feature, zoom int) (tiles []tiles.Tile) {
	for _, s := range f.Geometry {
		tiles = append(tiles, ShapeTiles(s, zoom)...)
	}
	return
}

// Shape tiles returns a list of tiles that cover a shape at the given zoom level
func ShapeTiles(shp *geo.Shape, zoom int) (tiles []tiles.Tile) {
	bb := shp.BoundingBox()
	ne := bb.NorthEast().ToTile(zoom)
	sw := bb.SouthWest().ToTile(zoom)
	cur := sw
	for x := sw.X; x <= ne.X; x++ {
		for y := sw.Y; y >= ne.Y; y-- { //origin is NW
			cur.X, cur.Y = x, y
			tiles = append(tiles, cur)
		}
	}
	return
}

// MvtAdapter populates an mvt feature from a diglet geo feature
func MvtAdapter(f *geo.Feature, t tiles.Tile) (a *mvt.Feature) {
	g := f.Type
	if g == geo.LineFeature {
		g = "linestring"
	}
	var id uint64
	id, _ = util.CastUInt64(f.ID)
	a = mvt.NewFeatureAdapter(&id, g, f.Properties)
	shps := tiledShapes(f, t)
	a.AddShape(shps...)
	return
}

func tiledShapes(f *geo.Feature, t tiles.Tile) (shps []*mvt.Shape) {
	//shps = make([]*mvt.Shape, len(f.Geometry))
	for _, s := range f.Geometry {
		shp := tiledShape(s, t)
		if shp.Len() > 0 {
			shps = append(shps, shp)
		}
		//shps[i] = shp
	}
	return
}

func tiledShape(gs *geo.Shape, t tiles.Tile) *mvt.Shape {
	s := make([]mvt.Point, len(gs.Coordinates))
	for i, c := range gs.Coordinates {
		pixel := tiles.ClippedCoords(c.Lat, c.Lon).ToPixel(t.Z)
		x := int(pixel.X - t.ToPixel().X)
		y := int(pixel.Y - t.ToPixel().Y)
		p := mvt.Point{X: x, Y: y}
		s[i] = p
	}
	shp := mvt.NewPolygon(s...) //could be line or point
	box, _ := mvt.NewBox(
		mvt.Point{-1 * ClipBuffer, -1 * ClipBuffer},
		mvt.Point{tiles.TileSize + ClipBuffer, tiles.TileSize + ClipBuffer},
	)
	if len(s) > 2 {
		box.Clip(shp)
	}
	return shp
}
