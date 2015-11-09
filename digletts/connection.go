// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package digletts

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	events chan TsEvent
	tiles  map[TileXYZ]uint
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) listen() *CodedError {
	go c.subscribe()
	defer c.ws.Close()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var req RequestMessage
		if err := c.ws.ReadJSON(&req); err != nil {
			// TODO figure out if this bloats from nor receiving close
			// Should this close the connection?
			warn(err, "readjson error")
			cerr := cerrorf(RpcInvalidRequest, err.Error())
			c.respond(cerr.ResponseMessage())
			return cerr
			//return cerrorf(400, err.Error())
		} else if cerr := req.Validate(); cerr != nil {
			c.respond(cerr.ResponseMessage())
			//return cerr
		} else {
			req.Params["wsconn"] = c
			if msg := req.ExecuteMethod(); msg != nil {
				warn(c.respond(msg), "conn respond")
			}
		}
	}
}

// Format, vals will be sprintf'd
func (c *connection) notify(format string, vals ...interface{}) error {
	msg := sprintf(format, vals...)
	return c.respond(SuccessMsg(msg))
}

func (c *connection) respond(msg *ResponseMessage) error {
	if payload, err := msg.Marshal(); err != nil {
		return err
	} else {
		c.write(websocket.TextMessage, payload)
	}
	return nil
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) subscribe() {
	//subscribe to tileset channel...
	//hub.subscribe <- c
	pinger := time.NewTicker(pingPeriod)
	defer func() {
		pinger.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case e, ok := <-c.events:
			if !ok {
				//tileset was closed message
				//c.write(websocket.CloseMessage, []byte{})
				info("I'm too lazy to raise an actual error, but the channel was closed")
				return
			}
			info("I'm too lazy to handle this event right now %s", e)
			//send all tiles
			//if err := c.write(websocket.TextMessage, payload); err != nil {
		case <-pinger.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				warn(err, "ping fucked")
				return
			}
		}
	}
}

func (c *connection) bindTile(xyz TileXYZ, reqId uint) {
	//TODO if concurrent access, use mutex
	c.tiles[xyz] = reqId
}

func (c *connection) unbindTile(xyz TileXYZ) {
	//TODO if concurrent access, use mutex
	delete(c.tiles, xyz)
}

func NewConnection(ws *websocket.Conn) *connection {
	return &connection{
		events: make(chan TsEvent),
		tiles:  make(map[TileXYZ]uint),
		ws:     ws,
	}
}
