package mvt

import (
	vt "github.com/buckhx/diglet/mbt/mvt/vector_tile"
	_ "github.com/buckhx/diglet/util"
)

type TileAdapter struct {
	X, Y, Z int
	tile    *vt.Tile
	//layers  map[string]*layer
}

func NewTileAdapter(x, y, z int) *TileAdapter {
	return &TileAdapter{
		X:    x,
		Y:    y,
		Z:    z,
		tile: &vt.Tile{},
		//layers: make(map[string]*layer),
	}
}

func (t *TileAdapter) NewLayer(name string, extent int) (layer *Layer) {
	layer = newLayer(name, extent)
	t.tile.Layers = append(t.tile.Layers, layer.vt_layer)
	return
}

func (t *TileAdapter) GetTile() *vt.Tile {
	return t.tile
}

func (t *TileAdapter) GetTileBytes() ([]byte, error) {
	return vt.Encode(t.GetTile())
}

func (t *TileAdapter) GetTileGz() ([]byte, error) {
	return vt.EncodeGzipped(t.GetTile())
}
