package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type deleter struct {
	DB store.Deleter
}

// NewDeleter takes the database dependency and uses it to implement
// the sidebar.Deleter interface. This interface is used to remove
// objects from store the database.
func NewDeleter(db store.Deleter) (sidebar.Deleter, error) {
	return &deleter{
		DB: db,
	}, nil
}

func (a *deleter) DeleteChannel(id string) (*sidebar.Channel, error) {
	return a.DB.DeleteChannel(id)
}

func (a *deleter) DeleteUser(id string) (*sidebar.User, error) {
	return a.DB.DeleteUser(id)
}
