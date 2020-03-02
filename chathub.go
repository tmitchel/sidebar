package sidebar

type chathub struct {
	clients    map[*client]bool
	broadcast  chan WebSocketMessage
	register   chan *client
	unregister chan *client
}

func NewChathub() *chathub {
	return &chathub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan WebSocketMessage),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}

func (h *chathub) run() {
	for {
		select {
		case client := <-h.register:
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

func (h *chathub) Unregister(cl *client) {
	h.unregister <- cl
}
