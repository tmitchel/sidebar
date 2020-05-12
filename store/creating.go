package store

import (
	"github.com/tmitchel/sidebar"
)

// Creater ...
type Creater interface {
	CreateUser(*sidebar.User) (*sidebar.User, error)
	CreateWorkspace(*sidebar.Workspace) (*sidebar.Workspace, error)
	CreateDefaultWorkspace(*sidebar.Workspace) (*sidebar.Workspace, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.ChatMessage) (*sidebar.ChatMessage, error)
}

// CreateUserNoToken is used to create a default user when the app starts
// and no token is available.
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

func (d *database) CreateUser(u *sidebar.User) (*sidebar.User, error) {
	_, err := psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(u.ID, u.DisplayName, u.Email, u.Password, u.ProfileImg).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (d *database) CreateWorkspace(w *sidebar.Workspace) (*sidebar.Workspace, error) {
	_, err := psql.Insert("workspaces").
		Columns("id", "display_name", "display_image", "token", "default_ws").
		Values(w.ID, w.DisplayName, w.DisplayImg, w.Token, false).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (d *database) CreateDefaultWorkspace(w *sidebar.Workspace) (*sidebar.Workspace, error) {
	_, err := psql.Insert("workspaces").
		Columns("id", "display_name", "display_image", "token", "default_ws").
		Values(w.ID, w.DisplayName, w.DisplayImg, w.Token, true).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return w, nil
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

func (d *database) CreateMessage(m *sidebar.ChatMessage) (*sidebar.ChatMessage, error) {
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
