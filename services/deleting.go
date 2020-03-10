package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type deleter struct {
	DB store.Deleter
}

func NewDeleter(db store.Deleter) (sidebar.Deleter, error) {
	return &deleter{
		DB: db,
	}, nil
}

func (a *deleter) DeleteChannel(id int) (*sidebar.Channel, error) {
	return a.DB.DeleteChannel(id)
}

func (a *deleter) DeleteUser(id int) (*sidebar.User, error) {
	return a.DB.DeleteUser(id)
}
