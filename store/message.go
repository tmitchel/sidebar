package store

import "github.com/tmitchel/sidebar"

type webSocketMessage struct {
	ID      int    `db:"id"`
	Event   int64  `db:"event"`
	Content string `db:"content"`
}

func MessageFromModel(m *sidebar.WebSocketMessage) *webSocketMessage {
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
