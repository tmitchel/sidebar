package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
)

func TestChannelFromModel(t *testing.T) {
	m := &sidebar.Channel{
		ID:   1,
		Name: "test-channel",
	}

	n := channelFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Name, n.Name)
}

func TestChannelToModel(t *testing.T) {
	n := &channel{
		ID:   1,
		Name: "test-channel",
	}

	m := n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Name, m.Name)
}
