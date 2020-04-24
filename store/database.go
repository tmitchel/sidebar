package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"

	_ "github.com/lib/pq" // postgres drivers
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Creater ...
type Creater interface {
	NewToken(string, string) error
	CreateUser(*sidebar.User, string) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
}

// Adder ...
type Adder interface {
	AddUserToChannel(string, string) error
	RemoveUserFromChannel(string, string) error
	ResolveChannel(string) error
}

// Deleter ...
type Deleter interface {
	DeleteUser(string) (*sidebar.User, error)
	DeleteChannel(string) (*sidebar.Channel, error)
}

// Getter ...
type Getter interface {
	GetUser(string) (*sidebar.User, error)
	GetChannel(string) (*sidebar.Channel, error)
	GetMessage(string) (*sidebar.WebSocketMessage, error)

	GetUsers() ([]*sidebar.User, error)
	GetChannels() ([]*sidebar.Channel, error)
	GetMessages() ([]*sidebar.WebSocketMessage, error)

	GetUsersInChannel(string) ([]*sidebar.User, error)
	GetChannelsForUser(string) ([]*sidebar.Channel, error)

	GetMessagesInChannel(string) ([]*sidebar.WebSocketMessage, error)
	GetMessagesFromUser(string) ([]*sidebar.WebSocketMessage, error)
	GetMessagesToUser(string) ([]*sidebar.WebSocketMessage, error)
}

// Updater ...
type Updater interface {
	UpdateUserInformation(*sidebar.User) error
	UpdateUserPassword(string, []byte) error
}

// Authenticater ...
type Authenticater interface {
	UserForAuth(string) (*sidebar.User, error)
}

// Database provides methods to query the database.
type Database interface {
	Adder
	Deleter
	Creater
	Getter
	Updater
	Authenticater
	sq.BaseRunner

	Close()
}

type database struct {
	*sql.DB
}

// New connects to the postgres database
// and returns that connection.
func New(psqlInfo string) (Database, error) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening database")
	}

	// make sure we have a good connection
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "Error pinging database")
	}

	return &database{db}, nil
}

// Close closes the database.
func (d *database) Close() {
	d.DB.Close()
}

func (d *database) NewToken(token string, userID string) error {
	_, err := psql.Insert("tokens").
		Columns("token", "creater_id").
		Values(token, userID).
		RunWith(d).Exec()

	return err
}

func CreateUserNoToken(d Database, u *sidebar.User) (*sidebar.User, error) {
	duser := userFromModel(u)
	_, err := psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(duser.ID, duser.DisplayName, duser.Email, duser.Password, "https://randomuser.me/api/portraits/women/81.jpg").
		Suffix("RETURNING id").
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return duser.ToModel(), nil
}

func createUser(d Database, u *sidebar.User) (*sidebar.User, error) {
	duser := userFromModel(u)
	_, err := psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(duser.ID, duser.DisplayName, duser.Email, duser.Password, "https://randomuser.me/api/portraits/women/81.jpg").
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return duser.ToModel(), nil
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

	duser := userFromModel(u)
	_, err = psql.Insert("users").
		Columns("id", "display_name", "email", "password", "profile_image").
		Values(duser.ID, duser.DisplayName, duser.Email, duser.Password, duser.ProfileImg).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	_, err = psql.Update("tokens").
		Set("valid", false).Set("new_user_id", duser.ID).
		Where(sq.Eq{"token": token}).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	return duser.ToModel(), nil
}

func (d *database) ResolveChannel(channelID string) error {
	c, _ := d.GetChannel(channelID)
	_, err := psql.Update("channels").
		Set("resolved", !c.Resolved).
		Where(sq.Eq{"id": channelID}).
		RunWith(d).Exec()
	return err
}

func (d *database) AddUserToChannel(userID, channelID string) error {
	_, err := psql.Insert("users_channels").
		Columns("user_id", "channel_id").Values(userID, channelID).
		RunWith(d).Exec()
	return err
}

func (d *database) RemoveUserFromChannel(userID, channelID string) error {
	_, err := psql.Delete("users_channels").
		Where(sq.Eq{"user_id": userID, "channel_id": channelID}).
		RunWith(d).Exec()
	return err
}

func (d *database) GetUser(id string) (*sidebar.User, error) {
	var u user
	row := psql.Select("id", "display_name", "email", "password", "profile_image").
		From("users").Where(sq.Eq{"id": id}).RunWith(d).QueryRow()
	err := row.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password, &u.ProfileImg)
	if err != nil {
		return nil, err
	}

	return u.ToModel(), nil
}

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
		var u user
		err := rows.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password, &u.ProfileImg)
		if err != nil {
			continue
		}
		users = append(users, u.ToModel())
	}

	return users, nil
}

func (d *database) GetChannelsForUser(id string) ([]*sidebar.Channel, error) {
	var parent sql.NullString
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		Join("users_channels uc ON ( uc.channel_id = ch.id )").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		Where(sq.Eq{"uc.user_id": id}).RunWith(d).Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c channel
		err := rows.Scan(&c.ID, &c.Name, &c.Details, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
		if err != nil {
			continue
		}

		if parent.Valid {
			c.Parent = parent.String
		}

		channels = append(channels, c.ToModel())
	}

	return channels, nil
}

func (d *database) UserForAuth(email string) (*sidebar.User, error) {
	var authUser userForAuth
	row := psql.Select("id", "password").
		From("users").Where(sq.Eq{"email": email}).RunWith(d).QueryRow()
	err := row.Scan(&authUser.ID, &authUser.Password)
	if err != nil {
		return nil, err
	}

	return authUser.ToModel(), nil
}

func (d *database) CreateChannel(c *sidebar.Channel) (*sidebar.Channel, error) {
	dchannel := channelFromModel(c)
	_, err := psql.Insert("channels").
		Columns("id", "display_name", "details", "is_sidebar", "is_direct").
		Values(dchannel.ID, dchannel.Name, dchannel.Details, dchannel.IsSidebar, dchannel.Direct).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	if dchannel.IsSidebar {
		_, err := psql.Insert("sidebars").
			Columns("id", "parent_id").Values(dchannel.ID, c.Parent).
			RunWith(d).Exec()
		if err != nil {
			return nil, err
		}
	}

	return dchannel.ToModel(), nil
}

func (d *database) GetChannel(id string) (*sidebar.Channel, error) {
	var parent sql.NullString
	var c channel
	row := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").Where(sq.Eq{"ch.id": id}).RunWith(d).QueryRow()
	err := row.Scan(&c.ID, &c.Name, &c.Details, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
	if err != nil {
		return nil, err
	}

	if parent.Valid {
		c.Parent = parent.String
	}

	return c.ToModel(), nil
}

func (d *database) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	dmessage := messageFromModel(m)
	_, err := psql.Insert("messages").
		Columns("id", "content", "event").Values(dmessage.ID, dmessage.Content, dmessage.Event).
		RunWith(d).Exec()

	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("users_messages").
		Columns("user_to_id", "user_from_id", "message_id").Values(m.ToUser, m.FromUser, dmessage.ID).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("channels_messages").
		Columns("channel_id", "message_id").Values(m.Channel, dmessage.ID).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	mod := dmessage.ToModel()
	mod.ToUser = m.ToUser
	mod.FromUser = m.FromUser
	mod.Channel = m.Channel
	return mod, nil
}

func (d *database) GetMessage(id string) (*sidebar.WebSocketMessage, error) {
	var m webSocketMessage
	var toUser, fromUser, channel string
	row := psql.Select("ms.id", "ms.event", "ms.content", "cm.channel_id", "um.user_from_id", "um.user_to_id").From("messages as ms").
		Join("channels_messages cm ON (cm.message_id = ms.id)").
		Join("users_messages um ON ( um.message_id = ms.id )").
		RunWith(d).QueryRow()
	err := row.Scan(&m.ID, &m.Event, &m.Content, &channel, &fromUser, &toUser)
	if err != nil {
		return nil, err
	}
	mod := m.ToModel()
	mod.Channel = channel
	mod.FromUser = fromUser
	mod.ToUser = toUser

	return mod, nil
}

func (d *database) GetMessagesInChannel(id string) ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "content", "event", "cm.channel_id", "um.user_from_id", "um.user_to_id").From("messages").
		Join("channels_messages cm ON ( cm.message_id = id )").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"cm.channel_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m webSocketMessage
		var channel, toUser, fromUser string
		err := rows.Scan(&m.ID, &m.Content, &m.Event, &channel, &fromUser, &toUser)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		mod := m.ToModel()
		mod.Channel = channel
		mod.ToUser = toUser
		mod.FromUser = fromUser

		messages = append(messages, mod)
	}

	return messages, nil
}

func (d *database) GetMessagesFromUser(id string) ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"um.user_from_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m webSocketMessage
		err := rows.Scan(&m.ID, &m.Content)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		messages = append(messages, m.ToModel())
	}

	return messages, nil
}

func (d *database) GetMessagesToUser(id string) ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("users_messages um ON ( um.message_id = id )").
		Where(sq.Eq{"um.user_to_id": id}).
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var m webSocketMessage
		err := rows.Scan(&m.ID, &m.Content)
		if err != nil {
			return nil, errors.New("Error scanning for message")
		}

		messages = append(messages, m.ToModel())
	}

	return messages, nil
}

func (d *database) GetUsers() ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "email", "display_name", "password", "profile_image").From("users").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any users")
	}

	for rows.Next() {
		var u user
		err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Password, &u.ProfileImg)
		if err != nil {
			return nil, errors.New("Error scanning users")
		}

		users = append(users, u.ToModel())
	}

	return users, nil
}

func (d *database) GetChannels() ([]*sidebar.Channel, error) {
	var parent sql.NullString
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.details", "ch.is_sidebar", "sb.parent_id", "ch.is_direct", "ch.resolved").
		From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.Errorf("Unable to find any channels %v", err)
	}

	for rows.Next() {
		var c channel
		err := rows.Scan(&c.ID, &c.Name, &c.Details, &c.IsSidebar, &parent, &c.Direct, &c.Resolved)
		if err != nil {
			return nil, errors.Errorf("Error scanning channels %v", err)
		}

		if parent.Valid {
			c.Parent = parent.String
		}

		channels = append(channels, c.ToModel())
	}

	return channels, nil
}

func (d *database) GetMessages() ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("ms.id", "ms.event", "ms.content", "um.user_to_id", "um.user_from_id", "cm.channel_id").
		From("messages as ms").
		Join("users_messages um ON (um.message_id = id)").
		Join("channels_messages cm ON (cm.message_id = id)").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var w webSocketMessage
		var toID, fromID, channelID string
		err := rows.Scan(&w.ID, &w.Event, &w.Content, &toID, &fromID, &channelID)
		if err != nil {
			return nil, errors.New("Error scanning messages")
		}
		model := w.ToModel()
		model.ToUser = toID
		model.FromUser = fromID
		model.Channel = channelID

		messages = append(messages, model)
	}

	return messages, nil
}

func (d *database) UpdateUserInformation(u *sidebar.User) error {
	_, err := psql.Update("users").
		Set("display_name", u.DisplayName).
		Set("email", u.Email).
		Set("profile_image", u.ProfileImg).
		Where(sq.Eq{"id": u.ID}).
		RunWith(d).Exec()

	return err
}

func (d *database) UpdateUserPassword(id string, password []byte) error {
	_, err := psql.Update("users").
		Set("password", password).
		Where(sq.Eq{"id": id}).
		RunWith(d).Exec()

	return err
}

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
