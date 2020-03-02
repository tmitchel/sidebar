package store

import (
	"github.com/tmitchel/sidebar"
)

type channel struct {
	ID   int    `db:"id"`
	Name string `db:"display_name"`
}

func ChannelFromModel(c *sidebar.Channel) *channel {
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
