package main

import (
	"fmt"
	"github.com/buckhx/diglet/transform/mvt"
	"github.com/buckhx/diglet/transform/mvt/vector_tile"
	"github.com/buckhx/diglet/transform/tile_system"
	"io/ioutil"
)

func main() {
	tile := tile_system.Tile{Z: 12, X: 2124, Y: 1373}
	data, err := ioutil.ReadFile("12_2124_1373.pbf.gz")
	if err != nil {
		panic(err)
	}
	vtile, err := vector_tile.DecodeGzipped(data)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%+v\n", tile.Layers[0].Features[0].Geometry)
	for f, feature := range vtile.Layers[0].Features {
		geometry := mvt.GeometryFromVectorTile(feature.Geometry)
		for i, cmd := range geometry.ToCommands() {
			fmt.Printf("%v - %s\n", i, cmd)
		}
		fmt.Printf("Feature %d\n", f)
		for s, shape := range geometry.ToShapes() {
			fmt.Printf("Shape %d\n", s)
			for p, point := range shape.GetPoints() {
				offset := tile_system.Pixel{X: uint(point.X), Y: uint(point.Y)}
				coords := tile.ToPixelWithOffset(offset).ToCoords()
				fmt.Printf("\t%d - %v\n", p, coords)
			}
		}
	}
	fmt.Printf("%d\n", *vtile.Layers[0].Extent)
	//fmt.Printf("%+v\n", feature.Geometry)
	//fmt.Printf("%+v\n", tile.Layers[0].GetKeys())
	//fmt.Printf("%+v\n", tile.Layers[0].GetValues())
}
