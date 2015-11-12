// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diglet

import (
	dig "github.com/buckhx/diglet/burrow"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type TilesetTopic struct {
	name        string
	subscribers map[*dig.Connection]bool
	events      chan TsEvent
	subscribe   chan *dig.Connection
	unsubscribe chan *dig.Connection
	shut        chan struct{}
}

func (t *TilesetTopic) open() {
	for {
		select {
		case c := <-t.subscribe:
			t.subscribers[c] = true
		case c := <-t.unsubscribe:
			if _, ok := t.subscribers[c]; ok {
				delete(t.subscribers, c)
			}
		case e := <-t.events:
			//if event was REMOVE, shut topic
			for c := range t.subscribers {
				//c.events <- e
				info("%s -> %s", e, c)
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
		subscribers: make(map[*dig.Connection]bool),
		events:      make(chan TsEvent),
		subscribe:   make(chan *dig.Connection),
		unsubscribe: make(chan *dig.Connection),
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
}

func (h *IoHub) publish(events <-chan TsEvent) {
	for event := range events {
		info("Tileset Change - %s", event.String())
		if topic, ok := h.topics[event.Name]; ok {
			topic.events <- event
		} else {
			info("Topic did not exist in hub for TsEvent: %s", event)
		}
	}
}

func (h *IoHub) subscribe(tilesetSlug string, conn *dig.Connection) (err error) {
	if topic, ok := h.topics[tilesetSlug]; ok {
		topic.subscribe <- conn
	} else {
		err = errorf("Topic does not exist for tileset %s", tilesetSlug)
	}
	return
}

func (h *IoHub) unsubscribe(tilesetSlug string, conn *dig.Connection) (err error) {
	if topic, ok := h.topics[tilesetSlug]; ok {
		topic.unsubscribe <- conn
	} else {
		err = errorf("Topic does not exist for tileset %s", tilesetSlug)
	}
	return
}

func NewHub(tilesets *TilesetIndex) (h *IoHub) {
	h = &IoHub{
		tilesets: tilesets,
		topics:   make(map[string]*TilesetTopic),
	}
	for slug, _ := range h.tilesets.Tilesets {
		h.topics[slug] = newTilesetTopic(slug)
	}
	return
}
