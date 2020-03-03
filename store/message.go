package store

import "github.com/tmitchel/sidebar"

type webSocketMessage struct {
	ID      int    `db:"id"`
	Event   int64  `db:"event"`
	Content string `db:"content"`
}

// messageFromModel converts the normal sidebar.WebSocketMessage model
// into a webSocketMessage which has properties only useful for the
// database.
func messageFromModel(m *sidebar.WebSocketMessage) *webSocketMessage {
	return &webSocketMessage{
		ID:      m.ID,
		Event:   m.Event,
		Content: m.Content,
	}
}

func (m *webSocketMessage) ToModel() *sidebar.WebSocketMessage {
	return &sidebar.WebSocketMessage{
		ID:      m.ID,
		Event:   m.Event,
		Content: m.Content,
	}
}
