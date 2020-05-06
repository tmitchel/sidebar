package store

import sq "github.com/Masterminds/squirrel"

// Adder provides methods for updating an existing channel. "Adder"
// probably isn't the right name anymore.
type Adder interface {
	AddUserToChannel(string, string) error
	RemoveUserFromChannel(string, string) error
	ResolveChannel(string) error
}

// AddUserToChannel takes a user id and channel id then adds that pair to
// the users_channels table.
func (d *database) AddUserToChannel(userID, channelID string) error {
	_, err := psql.Insert("users_channels").
		Columns("user_id", "channel_id").Values(userID, channelID).
		RunWith(d).Exec()
	return err
}

// RemoveUserFromChannel removes a user/channel pair from the users_channels table.
func (d *database) RemoveUserFromChannel(userID, channelID string) error {
	_, err := psql.Delete("users_channels").
		Where(sq.Eq{"user_id": userID, "channel_id": channelID}).
		RunWith(d).Exec()
	return err
}

// ResolveChannel sets the channel's Resolved field to the opposite of it's current
// value.
func (d *database) ResolveChannel(channelID string) error {
	c, _ := d.GetChannel(channelID)
	_, err := psql.Update("channels").
		Set("resolved", !c.Resolved).
		Where(sq.Eq{"id": channelID}).
		RunWith(d).Exec()
	return err
}
