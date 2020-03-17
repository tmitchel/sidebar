package frontend

import "github.com/tmitchel/sidebar"

type deleter struct {
}

func NewDeleter() sidebar.Deleter {
	return &deleter{}
}

func (a *deleter) DeleteChannel(id int) (*sidebar.Channel, error) {
	return nil, nil
}

func (a *deleter) DeleteUser(id int) (*sidebar.User, error) {
	return nil, nil
}
