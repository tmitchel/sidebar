package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
)

func TestMessageFromModel(t *testing.T) {
	m := &sidebar.WebSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	n := messageFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Event, n.Event)
	assert.Equal(t, m.Content, n.Content)
}

func TestMessageToModel(t *testing.T) {
	n := &webSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	m := n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Event, m.Event)
	assert.Equal(t, n.Content, m.Content)
}
