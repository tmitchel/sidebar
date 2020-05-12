package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/tmitchel/sidebar"
)

// Updater provides methods for updating rows stored in teh database.
type Updater interface {
	UpdateWorkspaceImage(string, string) error
	UpdateUserInformation(*sidebar.User) error
	UpdateChannelInformation(*sidebar.Channel) error
	UpdateUserPassword(string, []byte) error
}

// UpdateWorkspaceImage updates the image associated with the given
// workspace.
func (d *database) UpdateWorkspaceImage(id, img string) error {
	_, err := psql.Update("workspaces").
		Set("display_image", img).
		Where(sq.Eq{"id": id}).
		RunWith(d).Exec()
	return err
}

// UpdateUserInformation updates all information for the user EXCEPT
// the password. Pass the current value if you don't want to update a
// field.
func (d *database) UpdateUserInformation(u *sidebar.User) error {
	_, err := psql.Update("users").
		Set("display_name", u.DisplayName).
		Set("email", u.Email).
		Set("profile_image", u.ProfileImg).
		Where(sq.Eq{"id": u.ID}).
		RunWith(d).Exec()
	return err
}

// UpdateUserPassword sets the password for the given user to a new value.
func (d *database) UpdateUserPassword(id string, password []byte) error {
	_, err := psql.Update("users").
		Set("password", password).
		Where(sq.Eq{"id": id}).
		RunWith(d).Exec()
	return err
}

// UpdateChannelInformation updates all information for the channel. Pass
// the current value for any fields you don't want to update.
func (d *database) UpdateChannelInformation(c *sidebar.Channel) error {
	_, err := psql.Update("channels").
		Set("display_name", c.Name).
		Set("details", c.Details).
		Set("display_image", c.Image).
		Where(sq.Eq{"id": c.ID}).
		RunWith(d).Exec()
	return err
}
