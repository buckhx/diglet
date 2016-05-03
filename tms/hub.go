package tms

import (
	dig "github.com/buckhx/diglet/burrow"
)

type subMsg struct {
	conn *dig.Connection
	xyz  TileXYZ
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type TilesetTopic struct {
	name        string
	subscribers map[*dig.Connection]*tileSubscription
	events      chan TsEvent
	subscribe   chan *subMsg
	unsubscribe chan *subMsg
	shut        chan struct{}
}

func (t *TilesetTopic) open() {
	for {
		select {
		case m := <-t.subscribe:
			if _, ok := t.subscribers[m.conn]; !ok {
				t.subscribers[m.conn] = newTileSubscription()
			}
			t.subscribers[m.conn].add(m.xyz)
		case m := <-t.unsubscribe:
			if _, ok := t.subscribers[m.conn]; ok {
				t.subscribers[m.conn].remove(m.xyz)
				if !t.subscribers[m.conn].isEmpty() {
					delete(t.subscribers, m.conn)
				}
			}
		case e := <-t.events:
			func(e TsEvent) {}(e)
			for c, s := range t.subscribers {
				go s.notify(c)
				//info("ts notifying %s", e)
				//info("%s -> %s", e, c)
			}
		case <-t.shut:
			close(t.events)
			close(t.subscribe)
			close(t.unsubscribe)
		}
	}
}

func (t *TilesetTopic) close() {
	t.shut <- struct{}{}
}

func newTilesetTopic(name string) (topic *TilesetTopic) {
	topic = &TilesetTopic{
		name:        name,
		subscribers: make(map[*dig.Connection]*tileSubscription),
		events:      make(chan TsEvent),
		subscribe:   make(chan *subMsg),
		unsubscribe: make(chan *subMsg),
		shut:        make(chan struct{}),
	}
	return
}

type IoHub struct {
	tilesets *TilesetIndex
	topics   map[string]*TilesetTopic
}

func (h *IoHub) listen() {
	for _, topic := range h.topics {
		go topic.open()
	}
	go h.publish(h.tilesets.Events)
}

func (h *IoHub) publish(events <-chan TsEvent) {
	for event := range events {
		//info("Tileset Change - %s", event.String())
		//TODO remove/create messages
		//if event was REMOVE, shut topic
		if topic, ok := h.topics[event.Name]; ok {
			topic.events <- event
		} else {
			info("Topic did not exist in hub for TsEvent: %s", event)
		}
	}
}

func (h *IoHub) bindTile(ctx *dig.RequestContext, xyz TileXYZ) (err error) {
	msg := &subMsg{conn: ctx.Connection, xyz: xyz}
	if topic, ok := h.topics[xyz.Tileset]; ok {
		topic.subscribe <- msg
	} else {
		err = errorf("Tileset does not exist %s", xyz.Tileset)
	}
	return
}

func (h *IoHub) unbindTile(ctx *dig.RequestContext, xyz TileXYZ) (err error) {
	msg := &subMsg{conn: ctx.Connection, xyz: xyz}
	if topic, ok := h.topics[xyz.Tileset]; ok {
		topic.unsubscribe <- msg
	} else {
		err = errorf("Tileset does not exist %s", xyz.Tileset)
	}
	return
}

func NewHub(tilesets *TilesetIndex) (h *IoHub) {
	h = &IoHub{
		tilesets: tilesets,
		topics:   make(map[string]*TilesetTopic),
	}
	for slug := range h.tilesets.Tilesets {
		h.topics[slug] = newTilesetTopic(slug)
	}
	return
}

type tileSubscription struct {
	tiles map[TileXYZ]struct{}
}

func newTileSubscription() *tileSubscription {
	return &tileSubscription{
		tiles: make(map[TileXYZ]struct{}),
	}
}

func (s *tileSubscription) add(xyz TileXYZ) {
	s.tiles[xyz] = struct{}{}
}

func (s *tileSubscription) remove(xyz TileXYZ) {
	if _, ok := s.tiles[xyz]; !ok {
		delete(s.tiles, xyz)
	}
}

func (s *tileSubscription) isEmpty() bool {
	return len(s.tiles) == 0
}

func (s *tileSubscription) notify(conn *dig.Connection) {
	//TODO ops will have specific tile in the future?
	for xyz := range s.tiles {
		var msg *dig.ResponseMessage
		if tile, err := tilesets.read(xyz); err != nil {
			check(err)
			msg = dig.Cerrorf(dig.RpcInvalidRequest, err.Error()).ResponseMessage()
		} else {
			tile.Data = nil
			tile.Y = (1<<uint(tile.Z) - 1) - tile.Y
			id := sprintf("%d:%d:%d", tile.Z, tile.X, tile.Y)
			msg = dig.RespondMsg(id, tile)
		}
		conn.Respond(msg)
	}
}
