package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
)

// Getter provides methods for retrieiving different sets of
// data from the database.
type Getter interface {
	GetWorkspace(string) (*sidebar.Workspace, error)
	GetWorkspaceToken(string) (string, error)
	GetWorkspaceExists(string) error
	GetUser(string) (*sidebar.User, error)
	GetChannel(string) (*sidebar.Channel, error)
	GetMessage(string) (*sidebar.ChatMessage, error)

	GetUsers() ([]*sidebar.User, error)
	GetChannels() ([]*sidebar.Channel, error)
	GetMessages() ([]*sidebar.ChatMessage, error)

	GetUsersInChannel(string) ([]*sidebar.User, error)
	GetChannelsForUser(string) ([]*sidebar.Channel, error)

	GetMessagesInChannel(string) ([]*sidebar.ChatMessage, error)
	GetMessagesFromUser(string) ([]*sidebar.ChatMessage, error)
	GetMessagesToUser(string) ([]*sidebar.ChatMessage, error)
}

func (d *database) GetWorkspace(id string) (*sidebar.Workspace, error) {
	var w sidebar.Workspace
	err := psql.Select("id", "display_name", "display_image", "token").
		From("workspaces").Where(sq.Eq{"id": id}).RunWith(d).QueryRow().
		Scan(&w.ID, &w.DisplayName, &w.DisplayImg, &w.Token)
	if err != nil {
		return nil, err
	}
	return &w, err
}

func (d *database) GetWorkspaceToken(id string) (string, error) {
	var token string
	err := psql.Select("token").
		From("workspaces").Where(sq.Eq{"id": id}).RunWith(d).QueryRow().
		Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (d *database) GetWorkspaceExists(wid string) error {
	var id string
	err := psql.Select("id").From("workspaces").Where(sq.Eq{"id": wid}).
		RunWith(d).QueryRow().
		Scan(&id)
	if err != nil {
		return err
	}

	if id != wid {
		return errors.Errorf("Unable to locate workspace %v", wid)
	}

	return nil
}

// GetUser returns the user with the given id.
func (d *database) GetUser(id string) (*sidebar.User, error) {
	var u sidebar.User
	err := psql.Select("id", "display_name", "email", "password", "profile_image").
		From("users").Where(sq.Eq{"id": id}).RunWith(d).QueryRow().
		Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password, &u.ProfileImg)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUsersInChannel returns all users that are members of the given channel.
func (d *database) GetUsersInChannel(id string) ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "display_name", "email", "password", "profile_image").
		From("users").Join("users_channels uc ON ( uc.user_id = id )").
		Where(sq.Eq{"uc.channel_id": id}).RunWith(d).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u sidebar.User
		err := rows.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password, &u.ProfileImg)
		if err != nil {
			continue
		}
		users = append(users, &u)
	}

	return users, nil
}

// GetChannelsForUser returns all channels the given user is a member of.
func (d *database) GetChannelsForUser(id string) ([]*sidebar.Channel, error) {
	var parent sql.NullString
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.display_image", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		Join("users_channels uc ON ( uc.channel_id = ch.id )").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		Where(sq.Eq{"uc.user_id": id}).RunWith(d).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c sidebar.Channel
		err := rows.Scan(&c.ID, &c.Name, &c.Details, &c.Image, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
		if err != nil {
			continue
		}

		if parent.Valid {
			c.Parent = parent.String
		}

		channels = append(channels, &c)
	}

	return channels, nil
}

// GetChannel returns the channel with the given id.
func (d *database) GetChannel(id string) (*sidebar.Channel, error) {
	var parent sql.NullString
	var c sidebar.Channel
	err := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.display_image", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").Where(sq.Eq{"ch.id": id}).RunWith(d).QueryRow().
		Scan(&c.ID, &c.Name, &c.Details, &c.Image, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
	if err != nil {
		return nil, err
	}

	if parent.Valid {
		c.Parent = parent.String
	}

	return &c, nil
}

// GetMessage returns the message with the given id.
func (d *database) GetMessage(id string) (*sidebar.ChatMessage, error) {
	var m sidebar.ChatMessage
	err := psql.Select("ms.id", "ms.event", "ms.content", "cm.channel_id", "um.user_from_id", "um.user_to_id").From("messages as ms").
		Join("channels_messages cm ON (cm.message_id = ms.id)").
		Join("users_messages um ON ( um.message_id = ms.id )").
		RunWith(d).QueryRow().
		Scan(&m.ID, &m.Event, &m.Content, &m.Channel, &m.FromUser, &m.ToUser)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// GetMessageInChannel returns all messages sent in the given channel.
func (d *database) GetMessagesInChannel(id string) ([]*sidebar.ChatMessage, error) {
	var messages []*sidebar.ChatMessage
	rows, err := psql.Select("id", "content", "event", "cm.channel_id", "um.user_from_id", "um.user_to_id").From("messages").
		Join("channels_messages cm ON ( cm.message_id = id )").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"cm.channel_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m sidebar.ChatMessage
		err := rows.Scan(&m.ID, &m.Content, &m.Event, &m.Channel, &m.FromUser, &m.ToUser)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		messages = append(messages, &m)
	}

	return messages, nil
}

// GetMessagesFromUser returns all messages sent by the given user.
func (d *database) GetMessagesFromUser(id string) ([]*sidebar.ChatMessage, error) {
	var messages []*sidebar.ChatMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"um.user_from_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m sidebar.ChatMessage
		err := rows.Scan(&m.ID, &m.Content)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		messages = append(messages, &m)
	}

	return messages, nil
}

// GetMessagesToUser returns all messages sent to the given user.
func (d *database) GetMessagesToUser(id string) ([]*sidebar.ChatMessage, error) {
	var messages []*sidebar.ChatMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"um.user_to_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m sidebar.ChatMessage
		err := rows.Scan(&m.ID, &m.Content)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		messages = append(messages, &m)
	}

	return messages, nil
}

// GetUsers returns all users saved in the database.
func (d *database) GetUsers() ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "email", "display_name", "password", "profile_image").From("users").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any users")
	}

	for rows.Next() {
		var u sidebar.User
		err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Password, &u.ProfileImg)
		if err != nil {
			return nil, errors.New("Error scanning users")
		}

		users = append(users, &u)
	}

	return users, nil
}

// GetChannels returns all channels saved in the database.
func (d *database) GetChannels() ([]*sidebar.Channel, error) {
	var parent sql.NullString
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.display_image", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.Errorf("Unable to find any channels %v", err)
	}

	for rows.Next() {
		var c sidebar.Channel
		err := rows.Scan(&c.ID, &c.Name, &c.Details, &c.Image, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
		if err != nil {
			return nil, errors.Errorf("Error scanning channels %v", err)
		}

		if parent.Valid {
			c.Parent = parent.String
		}

		channels = append(channels, &c)
	}

	return channels, nil
}

// GetMessages returns all messages saved in the database.
func (d *database) GetMessages() ([]*sidebar.ChatMessage, error) {
	var messages []*sidebar.ChatMessage
	rows, err := psql.Select("ms.id", "ms.event", "ms.content", "um.user_to_id", "um.user_from_id", "cm.channel_id").
		From("messages as ms").
		Join("users_messages um ON (um.message_id = id)").
		Join("channels_messages cm ON (cm.message_id = id)").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var w sidebar.ChatMessage
		err := rows.Scan(&w.ID, &w.Event, &w.Content, &w.ToUser, &w.FromUser, &w.Channel)
		if err != nil {
			return nil, errors.New("Error scanning messages")
		}

		messages = append(messages, &w)
	}

	return messages, nil
}
