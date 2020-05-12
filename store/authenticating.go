package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
)

// Authenticater provies methods for authenticating a user with data
// from the database.
type Authenticater interface {
	UserForAuth(string) (*sidebar.User, error)
	UserInWorkspace(string, string) error
	ChannelInWorkspace(string, string) error
}

// UserForAuth takes a user email, queries the database for that user,
// and returns the user's id and password for authenticating.
func (d *database) UserForAuth(email string) (*sidebar.User, error) {
	var authUser sidebar.User
	err := psql.Select("id", "password").
		From("users").Where(sq.Eq{"email": email}).RunWith(d).QueryRow().
		Scan(&authUser.ID, &authUser.Password)
	if err != nil {
		return nil, err
	}

	return &authUser, nil
}

// UserInWorkspace returns an error if the user isn't a part
// of the provided workspace.
func (d *database) UserInWorkspace(uid, wid string) error {
	var id string
	err := psql.Select("user_id").From("workspaces_users").Where(sq.Eq{"workspace_id": wid}).Where(sq.Eq{"user_id": uid}).
		RunWith(d).QueryRow().Scan(&id)
	if err != nil {
		return err
	}

	if id != uid {
		return errors.Errorf("User %v not found", id)
	}

	return nil
}

// ChannelInWorkspace returns an error if the channel isn't a member
// of the workspace.
func (d *database) ChannelInWorkspace(cid, wid string) error {
	var id string
	err := psql.Select("channel_id").From("workspaces_channels").Where(sq.Eq{"workspace_id": wid}).Where(sq.Eq{"channel_id": cid}).
		RunWith(d).QueryRow().Scan(&id)
	if err != nil {
		return err
	}

	if id != cid {
		return errors.Errorf("User %v not found", id)
	}

	return nil
}
