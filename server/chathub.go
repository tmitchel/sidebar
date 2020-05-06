package server

import (
	"fmt"

	"github.com/tmitchel/sidebar"
)

type Storage interface {
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
}

type chathub struct {
	clients    map[*client]bool
	broadcast  chan sidebar.WebSocketMessage
	register   chan *client
	unregister chan *client
	db         Storage
}

// NewChathub creates a chathub to handle client Websocket
// connections and broadcasting messages.
func newChathub(s Storage) *chathub {
	return &chathub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan sidebar.WebSocketMessage),
		register:   make(chan *client),
		unregister: make(chan *client),
		db:         s,
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
