package server

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
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

type client struct {
	conn *websocket.Conn
	send chan sidebar.ChatMessage
	hub  *chathub
	User sidebar.User
}

// Digest decides how to handle a sidebar.ChatMessage based
// on the event type.
func (c *client) Digest(msg sidebar.ChatMessage) {
	switch msg.Event {
	case 1:
		// handle message
	case 2:
		// handle starting spin-off discussion
	}
}

// readPump listens for messages on the Websocket connection and
// send them to the chathub for broadcasting.
func (c *client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg sidebar.ChatMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
				logrus.Error("websocket closed by client")
			} else {
				logrus.Errorf("websocket error %v", err)
			}
			return
		}

		storedMsg, err := c.hub.db.CreateMessage(&msg)
		if err != nil {
			logrus.Errorf("error storing message %v", err)
		}
		if storedMsg != nil {
			c.hub.broadcast <- *storedMsg
		}
	}
}

// writePump listens for the chathub to broadcast a message then
// sends it out on the Websocket connection.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			if err := c.conn.WriteJSON(message); err != nil {
				logrus.Errorf("Error writing to websocket %v", err)
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logrus.Errorf("Error writing ping message. %v", err)
				return
			}
		}
	}
}
