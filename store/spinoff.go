package store

import "github.com/tmitchel/sidebar"

type spinoff struct {
	ID     int    `db:"id"`
	Name   string `db:"display_name"`
	Parent int    `db:"parent_id"`
}

// spinoffFromModel converts the normal sidebar.Spinoff model
// into a spinoff which has properties only useful for the
// database.
func spinoffFromModel(s *sidebar.Spinoff) *spinoff {
	return &spinoff{
		ID:     s.ID,
		Name:   s.Name,
		Parent: s.Parent,
	}
}

func (s *spinoff) ToModel() *sidebar.Spinoff {
	return &sidebar.Spinoff{
		ID:     s.ID,
		Name:   s.Name,
		Parent: s.Parent,
	}
}
