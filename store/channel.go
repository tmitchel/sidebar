package store

import (
	"github.com/tmitchel/sidebar"
)

type channel struct {
	ID   int    `db:"id"`
	Name string `db:"display_name"`
}

// channelFromModel converts the normal sidebar.Channel model
// into a channel which has properties only useful for the
// database.
func channelFromModel(c *sidebar.Channel) *channel {
	return &channel{
		ID:   c.ID,
		Name: c.Name,
	}
}

func (c *channel) ToModel() *sidebar.Channel {
	return &sidebar.Channel{
		ID:   c.ID,
		Name: c.Name,
	}
}
