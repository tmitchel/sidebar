package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type adder struct {
	DB store.Adder
}

// NewAdder takes the database dependency and uses it to implement
// the sidebar.Adder interface. This interface is used to create new
// objects and store them in the database.
func NewAdder(db store.Adder) (sidebar.Adder, error) {
	return &adder{
		DB: db,
	}, nil
}

func (a *adder) AddUserToChannel(userID, channelID string) error {
	return a.DB.AddUserToChannel(userID, channelID)
}

func (a *adder) RemoveUserFromChannel(userID, channelID string) error {
	return a.DB.RemoveUserFromChannel(userID, channelID)
}

func (a *adder) ResolveChannel(id string) error {
	return a.DB.ResolveChannel(id)
}
