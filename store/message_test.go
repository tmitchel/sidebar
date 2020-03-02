package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

func TestMessageFromModel(t *testing.T) {
	m := &sidebar.WebSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	n := store.MessageFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Event, n.Event)
	assert.Equal(t, m.Content, n.Content)
}

func TestMessageToModel(t *testing.T) {
	m := &sidebar.WebSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	n := store.MessageFromModel(m)
	m = n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Event, m.Event)
	assert.Equal(t, n.Content, m.Content)
}
