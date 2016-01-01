package main

import (
	"fmt"
	"github.com/buckhx/diglet/transform/mvt"
	"github.com/buckhx/diglet/transform/mvt/vector_tile"
	"io/ioutil"
)

func main() {
	data, err := ioutil.ReadFile("12_2124_1373.pbf.gz")
	if err != nil {
		panic(err)
	}
	tile, err := vector_tile.DecodeGzipped(data)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%+v\n", tile.Layers[0].Features[0].Geometry)
	for _, feature := range tile.Layers[0].Features {
		geometry := mvt.GeometryFromVectorTile(feature.Geometry)
		for i, cmd := range geometry.ToCommands() {
			fmt.Printf("%v - %s\n", i, cmd)
		}
	}
	fmt.Printf("%d\n", *tile.Layers[0].Extent)
	//fmt.Printf("%+v\n", feature.Geometry)
	//fmt.Printf("%+v\n", tile.Layers[0].GetKeys())
	//fmt.Printf("%+v\n", tile.Layers[0].GetValues())
}
