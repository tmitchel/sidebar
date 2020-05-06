package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
)

// Creater ...
type Creater interface {
	NewToken(string, string) error
	CreateUser(*sidebar.User, string) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
}

func (d *database) NewToken(token string, userID string) error {
	_, err := psql.Insert("tokens").
		Columns("token", "creater_id").
		Values(token, userID).
		RunWith(d).Exec()

	return err
}

func CreateUserNoToken(d Database, u *sidebar.User) (*sidebar.User, error) {
	_, err := psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(u.ID, u.DisplayName, u.Email, u.Password, "https://randomuser.me/api/portraits/women/81.jpg").
		Suffix("RETURNING id").
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return u, nil
}

func createUser(d Database, u *sidebar.User) (*sidebar.User, error) {
	_, err := psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(u.ID, u.DisplayName, u.Email, u.Password, "https://randomuser.me/api/portraits/women/81.jpg").
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (d *database) CreateUser(u *sidebar.User, token string) (*sidebar.User, error) {
	var valid bool
	err := psql.Select("valid").
		From("tokens").Where(sq.Eq{"token": token}).
		RunWith(d).QueryRow().Scan(&valid)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("Token is no longer valid")
	}

	_, err = psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(u.ID, u.DisplayName, u.Email, u.Password, u.ProfileImg).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	_, err = psql.Update("tokens").
		Set("valid", false).Set("new_user_id", u.ID).
		Where(sq.Eq{"token": token}).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (d *database) CreateChannel(c *sidebar.Channel) (*sidebar.Channel, error) {
	_, err := psql.Insert("channels").
		Columns("id", "display_name", "details", "display_image", "is_sidebar", "is_direct").
		Values(c.ID, c.Name, c.Details, c.Image, c.IsSidebar, c.Direct).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	if c.IsSidebar {
		_, err := psql.Insert("sidebars").
			Columns("id", "parent_id").Values(c.ID, c.Parent).
			RunWith(d).Exec()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (d *database) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	_, err := psql.Insert("messages").
		Columns("id", "content", "event").Values(m.ID, m.Content, m.Event).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("users_messages").
		Columns("user_to_id", "user_from_id", "message_id").Values(m.ToUser, m.FromUser, m.ID).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("channels_messages").
		Columns("channel_id", "message_id").Values(m.Channel, m.ID).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return m, nil
}
