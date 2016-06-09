package mbt

import (
	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/tiles"
)

type tileFeatures struct {
	t tiles.Tile
	f []*geo.Feature
}

func (tf tileFeatures) tileFeatures() (tiles.Tile, []*geo.Feature) {
	return tf.t, tf.f
}

type featureIndex struct {
	tiles tiles.TileIndex
}

func newFeatureIndex() (c *featureIndex) {
	return &featureIndex{
		tiles: tiles.NewSuffixIndex(),
	}
}

func (c *featureIndex) tileFeatures(zmin, zmax int) <-chan tileFeatures {
	tfs := make(chan tileFeatures, 1<<10)
	go func() {
		defer close(tfs)
		for t := range c.tiles.TileRange(zmin, zmax) {
			vals := c.tiles.Values(t)
			feats := make([]*geo.Feature, len(vals))
			for i, v := range vals {
				feats[i] = v.(*geo.Feature)
			}
			tf := tileFeatures{t: t, f: feats}
			tfs <- tf
		}
	}()
	return tfs
}

func (c *featureIndex) indexFeatures(features <-chan *geo.Feature, zoom int) {
	for f := range features {
		for _, t := range FeatureTiles(f, zoom) {
			c.tiles.Add(t, f)
		}
	}
}
