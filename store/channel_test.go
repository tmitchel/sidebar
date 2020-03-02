package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

func TestChannelFromModel(t *testing.T) {
	m := &sidebar.Channel{
		ID:   1,
		Name: "test-channel",
	}

	n := store.ChannelFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Name, n.Name)
}

func TestChannelToModel(t *testing.T) {
	m := &sidebar.Channel{
		ID:   1,
		Name: "test-channel",
	}

	n := store.ChannelFromModel(m)
	m = n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Name, m.Name)
}
