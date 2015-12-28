package main

import (
	"fmt"
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
	fmt.Printf("%+v\n", tile)
}
