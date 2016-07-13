package tms

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/buckhx/mbtiles"
	fsnotify "gopkg.in/fsnotify.v1"
)

type TsOp uint32

const (
	Upsert TsOp = 1 << iota
	Remove
)

const CacheName = ".diglet.cache"

type TsEvent struct {
	Name string
	Op   TsOp
}

func (tse *TsEvent) String() string {
	return fmt.Sprintf("%s %s", tse.Name, tse.Op)
}

// TilesetIndex is a container for tilesets loaded from disk
type TilesetIndex struct {
	Path     string
	Tilesets map[string]*mbtiles.Tileset
	Events   chan TsEvent
	watcher  *fsnotify.Watcher
	cache    *TileCache
}

// NewTilesetIndex creates a new tileset index, but does not read the tile tilesets from disk
func NewTilesetIndex(mbtilesDir string) (tsi *TilesetIndex, err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	watcher.Add(mbtilesDir)
	cache, err := InitTileCache(filepath.Join(mbtilesDir, CacheName))
	if err != nil {
		return
	}
	tsi = &TilesetIndex{
		Path:     mbtilesDir,
		Tilesets: make(map[string]*mbtiles.Tileset),
		Events:   make(chan TsEvent),
		watcher:  watcher,
		cache:    cache,
	}
	return
}

// ReadTilesets create a new tilesetindex and read the tilesets contents from disk
// It spawns goroutine that will refresh Tilesets from disk on changes
func ReadTilesets(mbtilesDir string) (tsi *TilesetIndex, err error) {
	tsi, err = NewTilesetIndex(mbtilesDir)
	if err != nil {
		return
	}
	mbtPaths, err := filepath.Glob(filepath.Join(mbtilesDir, "*.mbt*")) //match .mbtiles and mbt
	if err != nil {
		tsi = nil
		return
	}
	for _, path := range mbtPaths {
		if ts := tsi.readTileset(path); ts != nil {
			name := cleanTilesetName(path)
			tsi.Tilesets[name] = ts
		}
	}
	go tsi.watchMbtilesDir()
	return
}

// Read reads tile data from this index given an xyz
func (tsi *TilesetIndex) Read(xyz TileXYZ) (tile *mbtiles.Tile, err error) {
	if ts, ok := tsi.Tilesets[xyz.Tileset]; !ok {
		err = errorf("Tileset does not exist %s", xyz)
	} else {
		//tile, err = ts.ReadSlippyTile(xyz.X, xyz.Y, xyz.Z)
		// Again this is another hack to get try and wait until the DB is
		// done writing
		tile, ok = tsi.cache.GetTile(xyz.Tileset, xyz.String())
		if !ok {
			retries := 0
			retry := time.NewTicker(100 * time.Millisecond)
			for {
				select {
				case <-retry.C:
					tile, err = ts.ReadOSMTile(xyz.X, xyz.Y, xyz.Z)
					if err == nil {
						err = tsi.cache.PutTile(xyz.Tileset, xyz.String(), tile)
						check(err)
					}
					if err == nil || retries == 10 {
						return
					}
					warn(err, "ts read retry "+string(retries))
					retries += 1
				}
			}
		}
	}
	return
}

func (tsi *TilesetIndex) readTileset(path string) (ts *mbtiles.Tileset) {
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

func (tsi *TilesetIndex) watchMbtilesDir() {
	opBuf := newOpBuffer()
	go func() {
		for range opBuf.ticker.C {
			for _, op := range opBuf.flush() {
				info("fsnotify opbuffer flushed %s %d", op.String(), time.Now().UnixNano())
				var event TsEvent
				switch op.op {
				case fsnotify.Write:
					event = TsEvent{Op: Upsert, Name: op.tileset}
					go tsi.refreshCache(op.tileset)
				case fsnotify.Create:
					if ts := tsi.readTileset(op.tileset); ts != nil {
						tsi.Tilesets[op.tileset] = ts
					}
					event = TsEvent{Op: Upsert, Name: op.tileset}
				case fsnotify.Remove, fsnotify.Rename:
					if _, ok := tsi.Tilesets[op.tileset]; ok {
						delete(tsi.Tilesets, op.tileset)
					}
					event = TsEvent{Op: Remove, Name: op.tileset}
					go tsi.flushCache(op.tileset)
				default:
					continue
				}
				tsi.Events <- event
			}
		}
	}()
	for {
		select {
		case event := <-tsi.watcher.Events:
			//TODO make isMbtilesFile
			if !strings.HasSuffix(event.Name, ".mbtiles") && !strings.HasSuffix(event.Name, ".mbt") {
				continue
			}
			//info("fsnotify triggered %s", event.String())
			name := cleanTilesetName(event.Name)
			opBuf.add(event.Op, name)
		case err := <-tsi.watcher.Errors:
			warn(err, "fsnotify")
		}
	}
}

func (tsi *TilesetIndex) refreshCache(bucket string) {
	if ts, ok := tsi.Tilesets[bucket]; ok {
		tsi.cache.MapTiles(bucket, func(tile *mbtiles.Tile) (*mbtiles.Tile, error) {
			return ts.ReadOSMTile(tile.X, tile.Y, tile.Z)
		})
	}
}

func (tsi *TilesetIndex) flushCache(bucket string) {
	tsi.cache.DropBucket(bucket)
}

type TileXYZ struct {
	Tileset string `json:"tileset"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Z       int    `json:"z"`
}

func (xyz TileXYZ) String() string {
	if b, err := json.Marshal(xyz); err != nil {
		warn(err, "Could not marshal tile_xyz")
		return sprintf("Could not marshal tile_xyz %s", xyz)
	} else {
		return string(b)
	}
}

func fsnotifyOpString(op fsnotify.Op) string {
	s := "fsnotify."
	switch op {
	case fsnotify.Create:
		s += "Create"
	case fsnotify.Write:
		s += "Write"
	case fsnotify.Remove:
		s += "Remove"
	case fsnotify.Rename:
		s += "Rename"
	case fsnotify.Chmod:
		s += "Chmod"
	default:
		return "Unknown"
	}
	return s
}

type opBufOp struct {
	op      fsnotify.Op
	tileset string
}

func (op *opBufOp) String() string {
	return sprintf("{%s - %s}", op.tileset, fsnotifyOpString(op.op))
}

// fsnotify fires many events when a file is replaced
// this buffers those operations so that only one is fired
type opBuffer struct {
	sync.RWMutex
	ops    map[opBufOp]struct{}
	ticker *time.Ticker
}

func newOpBuffer() *opBuffer {
	return &opBuffer{
		ops:    make(map[opBufOp]struct{}),
		ticker: time.NewTicker(time.Millisecond * 200),
	}
}

func (b *opBuffer) add(op fsnotify.Op, tileset string) {
	b.Lock()
	b.ops[opBufOp{op: op, tileset: tileset}] = struct{}{}
	b.Unlock()
}

func (b *opBuffer) flush() []opBufOp {
	b.Lock()
	keys := make([]opBufOp, 0, len(b.ops))
	for k := range b.ops {
		keys = append(keys, k)
	}
	b.ops = make(map[opBufOp]struct{})
	b.Unlock()
	return keys
}
