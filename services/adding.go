package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type adder struct {
	DB store.Adder
}

func NewAdder(db store.Adder) (sidebar.Adder, error) {
	return &adder{
		DB: db,
	}, nil
}

func (a *adder) AddUserToChannel(userID, channelID int) error {
	return a.DB.AddUserToChannel(userID, channelID)
}

func (a *adder) RemoveUserFromChannel(userID, channelID int) error {
	return a.DB.RemoveUserFromChannel(userID, channelID)
}

func (a *adder) ResolveChannel(id int) error {
	return a.DB.ResolveChannel(id)
}
