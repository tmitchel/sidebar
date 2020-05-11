package server

import (
	"fmt"

	"github.com/tmitchel/sidebar"
)

type chathub struct {
	clients    map[*client]bool
	broadcast  chan sidebar.WebsocketMessage
	register   chan *client
	unregister chan *client
}

// NewChathub creates a chathub to handle client Websocket
// connections and broadcasting messages.
func newChathub() *chathub {
	return &chathub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan sidebar.WebsocketMessage),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}

func (h *chathub) run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("Registering client")
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
