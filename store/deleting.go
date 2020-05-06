package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
)

// Deleter provides methods for deleting rows from the database.
type Deleter interface {
	DeleteUser(string) (*sidebar.User, error)
	DeleteChannel(string) (*sidebar.Channel, error)
}

// DeleteUser removes the user with the given id from the database.
func (d *database) DeleteUser(id string) (*sidebar.User, error) {
	user, err := d.GetUser(id)
	if err != nil || user.ID == "" {
		return nil, errors.Errorf("User with id: %v doesn't exist", id)
	}

	_, err = psql.Delete("users").Where(sq.Eq{"id": id}).RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteChannel removes the channel with the given id from the database.
func (d *database) DeleteChannel(id string) (*sidebar.Channel, error) {
	channel, err := d.GetChannel(id)
	if err != nil || channel.ID == "" {
		return nil, errors.Errorf("User with id: %v doesn't exist", id)
	}

	_, err = psql.Delete("channels").Where(sq.Eq{"id": id}).RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return channel, nil
}
