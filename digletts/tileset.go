package digletts

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buckhx/mbtiles"
	"github.com/gorilla/mux"
	fsnotify "gopkg.in/fsnotify.v1"
)

var tilesets *TilesetIndex

func TilesetRoutes(prefix, mbtPath string) (r *RouteHandler) {
	tilesets = ReadTilesets(mbtPath)
	r = &RouteHandler{prefix, []Route{
		Route{"/{ts}/{z}/{x}/{y}", TileHandler},
		Route{"/{ts}", MetadataHandler},
		Route{"/", ListHandler},
	}}
	go func() {
		for event := range tilesets.Events {
			info("Tileset Change - %s", event.String())
		}
	}()
	return
}

// Reads the tile, dynamically determines enconding and content-type
func TileHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	vars := mux.Vars(r)
	tile, err := tilesets.tileFromVars(vars)
	if err != nil {
		return
	}
	headers := formatEncoding[tile.SniffFormat()]
	for _, h := range headers {
		w.Header().Set(h.key, h.value)
	}
	w.Header().Set("Content-Length", strconv.Itoa(binary.Size(tile.Data)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(tile.Data)
	return
}

// Get the metadatadata map from the tileset
func MetadataHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	//TODO if there's a json field, try to deserialze that
	vars := mux.Vars(r)
	slug := vars["ts"]
	if ts, ok := tilesets.Tilesets[slug]; ok {
		response = Success(ts.Metadata().Attributes())
	} else {
		response = Error(http.StatusBadRequest, fmt.Sprintf("No tileset named %q", slug))
	}
	return
}

// List the tilesets available on the server
func ListHandler(w http.ResponseWriter, r *http.Request) (response *JsonResponse) {
	tss := make(map[string]map[string]string)
	for name, ts := range tilesets.Tilesets {
		tss[name] = ts.Metadata().Attributes()
	}
	response = Success(tss)
	return
}

type TsOp uint32

const (
	Upsert TsOp = 1 << iota
	Remove
)

type TsEvent struct {
	Name string
	Op   TsOp
}

func (tse *TsEvent) String() string {
	return fmt.Sprintf("%s %s", tse.Name, tse.Op)
}

// Container for tilesets loaded from disk
type TilesetIndex struct {
	Path     string
	Tilesets map[string]*mbtiles.Tileset
	Events   chan TsEvent
	watcher  *fsnotify.Watcher
}

// Creates a new tileset index, but does not read the tile tilesets from disk
func NewTilesetIndex(mbtilesDir string) (tsi *TilesetIndex) {
	watcher, err := fsnotify.NewWatcher()
	check(err)
	watcher.Add(mbtilesDir)
	tsi = &TilesetIndex{
		Path:     mbtilesDir,
		Tilesets: make(map[string]*mbtiles.Tileset),
		Events:   make(chan TsEvent),
		watcher:  watcher,
	}
	return
}

// Create a new tilesetindex and read the tilesets contents from disk
// Spawns goroutine that will refresh TIlesets from disk on changes
func ReadTilesets(mbtilesDir string) (tsi *TilesetIndex) {
	tsi = NewTilesetIndex(mbtilesDir)
	mbtPaths, err := filepath.Glob(filepath.Join(mbtilesDir, "*.mbtiles"))
	check(err)
	readTileset := func(path string) (ts *mbtiles.Tileset) {
		ts, err := mbtiles.ReadTileset(path)
		if err != nil {
			warn(err, "skipping "+path)
			return
		}
		name := cleanTilesetName(path)
		if _, exists := tsi.Tilesets[name]; exists {
			check(fmt.Errorf("Multiple tilesets with slug %q like %q", name, path))
		}
		return

	}
	for _, path := range mbtPaths {
		if ts := readTileset(path); ts != nil {
			name := cleanTilesetName(path)
			tsi.Tilesets[name] = ts
		}
	}
	watchMbtilesDir := func() {
		for {
			select {
			case event := <-tsi.watcher.Events:
				//TODO make isMbtilesFile
				if !strings.HasSuffix(event.Name, ".mbtiles") {
					continue
				}
				info("fsnotify triggered %s", event.String())
				name := cleanTilesetName(event.Name)
				switch event.Op {
				case fsnotify.Write:
					tsi.Events <- TsEvent{Op: Upsert, Name: name}
				case fsnotify.Create:
					if ts := readTileset(event.Name); ts != nil {
						tsi.Tilesets[name] = ts
					}
					tsi.Events <- TsEvent{Op: Upsert, Name: name}
				case fsnotify.Remove, fsnotify.Rename:
					if _, ok := tsi.Tilesets[name]; ok {
						delete(tsi.Tilesets, name)
					}
					tsi.Events <- TsEvent{Op: Remove, Name: name}
				default:
					continue
				}
			case err := <-tsi.watcher.Errors:
				warn(err, "fsnotify")
			}
		}
	}
	go watchMbtilesDir()
	return
}

func (tsi *TilesetIndex) tileFromVars(vars map[string]string) (tile *mbtiles.Tile, err error) {
	slug := vars["ts"]
	x, err := strconv.Atoi(vars["x"])
	y, err := strconv.Atoi(vars["y"])
	z, err := strconv.Atoi(vars["z"])
	if ts, ok := tsi.Tilesets[slug]; ok && err == nil {
		tile, err = ts.ReadSlippyTile(x, y, z)
	} else {
		err = fmt.Errorf("No tileset with slug %q", slug)
	}
	return
}

func cleanTilesetName(path string) (slug string) {
	f := filepath.Base(path)
	f = strings.TrimSuffix(f, filepath.Ext(f))
	slug = slugged(f)
	return
}
