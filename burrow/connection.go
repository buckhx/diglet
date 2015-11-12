// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package burrow

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// connection is an middleman between the websocket connection and the hub.
type Connection struct {
	ws *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) listen(methods map[string]Method) *CodedError {
	go c.ping()
	defer c.ws.Close()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var req RequestMessage
		if err := c.ws.ReadJSON(&req); err != nil {
			cerr := cerrorf(RpcInvalidRequest, err.Error())
			c.respond(cerr.ResponseMessage())
			return cerr
		} else if cerr := req.Validate(); cerr != nil {
			c.respond(cerr.ResponseMessage())
		} else {
			ctx := &RequestContext{
				Request:    &req,
				Connection: c,
			}
			if method, ok := methods[ctx.Request.MethodName()]; !ok {
				msg := cerrorf(RpcMethodNotFound, "The method does not exist! %s", method).ResponseMessage()
				warn(c.respond(msg), "conn respond error")
			} else {
				if msg := method.Execute(ctx); msg != nil {
					warn(c.respond(msg), "conn respond error")
				}
			}
		}
	}
}

// Format, vals will be sprintf'd
func (c *Connection) notify(format string, vals ...interface{}) error {
	msg := sprintf(format, vals...)
	return c.respond(SuccessMsg(msg))
}

func (c *Connection) respond(msg *ResponseMessage) error {
	if payload, err := msg.Marshal(); err != nil {
		return err
	} else {
		c.write(websocket.TextMessage, payload)
	}
	return nil
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) ping() {
	//subscribe to tileset channel...
	//hub.subscribe <- c
	pinger := time.NewTicker(pingPeriod)
	defer func() {
		pinger.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case <-pinger.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{
		ws: ws,
	}
}
