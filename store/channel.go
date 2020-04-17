package store

import (
	"github.com/tmitchel/sidebar"
)

type channel struct {
	ID        string
	Name      string
	IsSidebar bool
	Parent    string
	Direct    bool
	Resolved  bool
}

// channelFromModel converts the normal sidebar.Channel model
// into a channel which has properties only useful for the
// database.
func channelFromModel(c *sidebar.Channel) *channel {
	return &channel{
		ID:        c.ID,
		Name:      c.Name,
		IsSidebar: c.IsSidebar,
		Parent:    c.Parent,
		Direct:    c.Direct,
		Resolved:  c.Resolved,
	}
}

func (c *channel) ToModel() *sidebar.Channel {
	return &sidebar.Channel{
		ID:        c.ID,
		Name:      c.Name,
		IsSidebar: c.IsSidebar,
		Parent:    c.Parent,
		Direct:    c.Direct,
		Resolved:  c.Resolved,
	}
}
