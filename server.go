package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func checks(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

type Tile struct {
	x, y, z int
}

func NewTile(x, y, z int) *Tile {
	return &Tile{x: x, y: y, z: z}
}

func TileFromVars(vars map[string]string) (tile *Tile, err error) {
	x, xerr := strconv.Atoi(vars["x"])
	y, xerr := strconv.Atoi(vars["y"])
	z, xerr := strconv.Atoi(vars["z"])
	tile = NewTile(x, y, z)
	err = xerr
	return
}

func TileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(request)
	tile, _ := TileFromVars(vars)
	fmt.Fprintf(out, "Request tile: %s\n", tile)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tile/{x}/{y}/{z}", TileHandler)

	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
