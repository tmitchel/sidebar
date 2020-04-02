package store

import (
	"database/sql"
	"fmt"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq" // postgres drivers
)

const (
	host     = "localhost"
	port     = 5432
	psqluser = "postgres"
	password = "example"
	dbname   = "sidebar"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Creater ...
type Creater interface {
	NewToken(string, int) error
	CreateUser(*sidebar.User, string) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
}

// Adder ...
type Adder interface {
	AddUserToChannel(int, int) error
	RemoveUserFromChannel(int, int) error
}

// Deleter ...
type Deleter interface {
	DeleteUser(int) (*sidebar.User, error)
	DeleteChannel(int) (*sidebar.Channel, error)
}

// Getter ...
type Getter interface {
	GetUser(int) (*sidebar.User, error)
	GetChannel(int) (*sidebar.Channel, error)
	GetMessage(int) (*sidebar.WebSocketMessage, error)

	GetUsers() ([]*sidebar.User, error)
	GetChannels() ([]*sidebar.Channel, error)
	GetMessages() ([]*sidebar.WebSocketMessage, error)

	GetUsersInChannel(int) ([]*sidebar.User, error)
	GetChannelsForUser(int) ([]*sidebar.Channel, error)

	GetMessagesInChannel(int) ([]*sidebar.WebSocketMessage, error)
	GetMessagesFromUser(int) ([]*sidebar.WebSocketMessage, error)
	GetMessagesToUser(int) ([]*sidebar.WebSocketMessage, error)
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
	Authenticater
	sq.BaseRunner

	Close()
}

type database struct {
	*sql.DB
}

// New connects to the postgres database
// and returns that connection.
func New() (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, psqluser, password, dbname)

	if os.Getenv("DATABASE_URL") != "" {
		psqlInfo = os.Getenv("DATABASE_URL")
	}
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

// NewWithMigration connects to the postgres database, runs migrations,
// and returns that connection.
func NewWithMigration(newDBName string) (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, psqluser, password, newDBName)

	if os.Getenv("DATABASE_URL") != "" {
		psqlInfo = os.Getenv("DATABASE_URL")
	}
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening database")
	}

	// make sure we have a good connection
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "Error pinging database")
	}

	if err := migrations(db); err != nil {
		return nil, err
	}

	return &database{db}, nil
}

// Close closes the database.
func (d *database) Close() {
	d.DB.Close()
}

func MigrationsForTesting(d Database) error {
	users := []*sidebar.User{
		&sidebar.User{
			ID:          1,
			DisplayName: "user one",
			Email:       "userone@email.com",
			Password:    []byte("password"),
		},
		&sidebar.User{
			ID:          2,
			DisplayName: "user twp",
			Email:       "usertwp@email.com",
			Password:    []byte("password"),
		},
		&sidebar.User{
			ID:          3,
			DisplayName: "user three",
			Email:       "userthree@email.com",
			Password:    []byte("password"),
		},
	}
	for i, u := range users {
		u.Password, _ = bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
		uu, err := createUser(d, u)
		if err != nil {
			return err
		}
		users[i].ID = uu.ID
	}

	channels := []*sidebar.Channel{
		&sidebar.Channel{
			ID:        1,
			Name:      "channel one",
			IsSidebar: false,
		},
		&sidebar.Channel{
			ID:        2,
			Name:      "channel two",
			IsSidebar: true,
			Parent:    1,
		},
	}
	for i, c := range channels {
		cc, err := d.CreateChannel(c)
		if err != nil {
			return err
		}
		channels[i].ID = cc.ID
	}

	messages := []*sidebar.WebSocketMessage{
		&sidebar.WebSocketMessage{
			ID:       1,
			Event:    1,
			Content:  "message one",
			ToUser:   users[0].ID,
			FromUser: users[1].ID,
			Channel:  channels[0].ID,
		},
		&sidebar.WebSocketMessage{
			ID:       2,
			Event:    1,
			Content:  "message two",
			ToUser:   users[1].ID,
			FromUser: users[0].ID,
			Channel:  channels[1].ID,
		},
		&sidebar.WebSocketMessage{
			ID:       3,
			Event:    2,
			Content:  "",
			ToUser:   users[0].ID,
			FromUser: users[0].ID,
			Channel:  channels[0].ID,
		},
	}
	for i, m := range messages {
		mm, err := d.CreateMessage(m)
		if err != nil {
			return err
		}
		messages[i].ID = mm.ID
	}

	for _, u := range users {
		err := d.AddUserToChannel(u.ID, channels[0].ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *database) NewToken(token string, userID int) error {
	_, err := psql.Insert("tokens").
		Columns("token", "creater_id").
		Values(token, userID).
		RunWith(d).Exec()

	return err
}

func createUser(d Database, u *sidebar.User) (*sidebar.User, error) {
	var id int
	duser := userFromModel(u)
	err := psql.Insert("users").
		Columns("display_name", "email", "password").
		Values(duser.DisplayName, duser.Email, duser.Password).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	duser.ID = id
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

	var id int
	duser := userFromModel(u)
	err = psql.Insert("users").
		Columns("display_name", "email", "password").
		Values(duser.DisplayName, duser.Email, duser.Password).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("tokens").
		Columns("valid", "new_user_id").Values(false, id).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	duser.ID = id
	return duser.ToModel(), nil
}

func (d *database) AddUserToChannel(userID, channelID int) error {
	_, err := psql.Insert("users_channels").
		Columns("user_id", "channel_id").Values(userID, channelID).
		RunWith(d).Exec()
	return err
}

func (d *database) RemoveUserFromChannel(userID, channelID int) error {
	_, err := psql.Delete("users_channels").
		Where(sq.Eq{"user_id": userID, "channel_id": channelID}).
		RunWith(d).Exec()
	return err
}

func (d *database) GetUser(id int) (*sidebar.User, error) {
	var u user
	row := psql.Select("id", "display_name", "email", "password").
		From("users").Where(sq.Eq{"id": id}).RunWith(d).QueryRow()
	err := row.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}

	return u.ToModel(), nil
}

func (d *database) GetUsersInChannel(id int) ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "display_name", "email", "password").
		From("users").Join("users_channels uc ON ( uc.user_id = id )").
		Where(sq.Eq{"uc.channel_id": id}).RunWith(d).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u user
		err := rows.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password)
		if err != nil {
			continue
		}
		users = append(users, u.ToModel())
	}

	return users, nil
}

func (d *database) GetChannelsForUser(id int) ([]*sidebar.Channel, error) {
	var parent sql.NullInt64
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.is_sidebar", "sb.parent_id", "ch.is_direct").From("channels as ch").
		Join("users_channels uc ON ( uc.channel_id = ch.id )").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		Where(sq.Eq{"uc.user_id": id}).RunWith(d).Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c channel
		err := rows.Scan(&c.ID, &c.Name, &c.IsSidebar, &parent, &c.Direct)
		if err != nil {
			continue
		}

		if parent.Valid {
			c.Parent = int(parent.Int64)
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
	var id int
	dchannel := channelFromModel(c)
	err := psql.Insert("channels").
		Columns("display_name", "is_sidebar", "is_direct").Values(dchannel.Name, dchannel.IsSidebar, dchannel.Direct).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	if dchannel.IsSidebar {
		_, err := psql.Insert("sidebars").
			Columns("id", "parent_id").Values(id, c.Parent).
			RunWith(d).Exec()
		if err != nil {
			return nil, err
		}
	}

	dchannel.ID = id
	return dchannel.ToModel(), nil
}

func (d *database) GetChannel(id int) (*sidebar.Channel, error) {
	var parent sql.NullInt64
	var c channel
	row := psql.Select("ch.id", "ch.display_name", "ch.is_sidebar", "sb.parent_id", "ch.is_direct").From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").Where(sq.Eq{"ch.id": id}).RunWith(d).QueryRow()
	err := row.Scan(&c.ID, &c.Name, &c.IsSidebar, &parent, &c.Direct)
	if err != nil {
		return nil, err
	}

	if parent.Valid {
		c.Parent = int(parent.Int64)
	}

	return c.ToModel(), nil
}

func (d *database) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	var id int
	dmessage := messageFromModel(m)
	err := psql.Insert("messages").
		Columns("content", "event").Values(dmessage.Content, dmessage.Event).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("users_messages").
		Columns("user_to_id", "user_from_id", "message_id").Values(m.ToUser, m.FromUser, id).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	_, err = psql.Insert("channels_messages").
		Columns("channel_id", "message_id").Values(m.Channel, id).
		RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	dmessage.ID = id
	return dmessage.ToModel(), nil
}

func (d *database) GetMessage(id int) (*sidebar.WebSocketMessage, error) {
	var m webSocketMessage
	var toUser, fromUser, channel int
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

func (d *database) GetMessagesInChannel(id int) ([]*sidebar.WebSocketMessage, error) {
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
		var channel, toUser, fromUser int
		err := rows.Scan(&m.ID, &m.Content, &m.Event, &channel, &toUser, &fromUser)
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

func (d *database) GetMessagesFromUser(id int) ([]*sidebar.WebSocketMessage, error) {
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

func (d *database) GetMessagesToUser(id int) ([]*sidebar.WebSocketMessage, error) {
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
	rows, err := psql.Select("id", "email", "display_name", "password").From("users").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any users")
	}

	for rows.Next() {
		var u user
		err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Password)
		if err != nil {
			return nil, errors.New("Error scanning users")
		}

		users = append(users, u.ToModel())
	}

	return users, nil
}

func (d *database) GetChannels() ([]*sidebar.Channel, error) {
	var parent sql.NullInt64
	var channels []*sidebar.Channel
	rows, err := psql.Select("ch.id", "ch.display_name", "ch.is_sidebar", "sb.parent_id", "ch.is_direct").From("channels as ch").
		JoinClause("FULL JOIN sidebars sb ON (sb.id = ch.id)").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.Errorf("Unable to find any channels %v", err)
	}

	for rows.Next() {
		var c channel
		err := rows.Scan(&c.ID, &c.Name, &c.IsSidebar, &parent, &c.Direct)
		if err != nil {
			return nil, errors.Errorf("Error scanning channels %v", err)
		}

		if parent.Valid {
			c.Parent = int(parent.Int64)
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
		var toID, fromID, channelID int
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

func (d *database) DeleteUser(id int) (*sidebar.User, error) {
	user, err := d.GetUser(id)
	if err != nil || user.ID == 0 {
		return nil, errors.Errorf("User with id: %v doesn't exist", id)
	}

	_, err = psql.Delete("users").Where(sq.Eq{"id": id}).RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *database) DeleteChannel(id int) (*sidebar.Channel, error) {
	channel, err := d.GetChannel(id)
	if err != nil || channel.ID == 0 {
		return nil, errors.Errorf("User with id: %v doesn't exist", id)
	}

	_, err = psql.Delete("channels").Where(sq.Eq{"id": id}).RunWith(d).Exec()
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func migrations(db *sql.DB) error {
	userQuery := `
	DROP TABLE IF EXISTS users CASCADE;
	CREATE TABLE users (
		id SERIAL UNIQUE,
		display_name TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		PRIMARY KEY(id)
	);`

	messageQuery := `
	DROP TABLE IF EXISTS messages CASCADE;
	CREATE TABLE messages (
		id SERIAL UNIQUE,
		content TEXT NOT NULL,
		event INT NOT NULL,
		PRIMARY KEY(id)
	);`

	channelQuery := `
	DROP TABLE IF EXISTS channels CASCADE;
	CREATE TABLE channels (
		id SERIAL UNIQUE,
		display_name TEXT UNIQUE NOT NULL,
		is_sidebar BOOLEAN DEFAULT FALSE,
		is_direct BOOLEAN DEFAULT FALSE,
		PRIMARY KEY(id)
	);`

	sidebarQuery := `
	DROP TABLE IF EXISTS sidebars;
	CREATE TABLE sidebars (
		id INT NOT NULL,
		parent_id INT,
		FOREIGN KEY(id) REFERENCES channels(id) ON DELETE CASCADE,
		FOREIGN KEY(parent_id) REFERENCES channels(id) ON DELETE CASCADE
	);`

	userChannelQuery := `
	DROP TABLE IF EXISTS users_channels CASCADE;
	CREATE TABLE users_channels (
		user_id INT REFERENCES users (id) ON UPDATE CASCADE,
		channel_id INT REFERENCES channels (id) ON UPDATE CASCADE,
		CONSTRAINT users_channels_pkey PRIMARY KEY (user_id, channel_id)
	);`

	tokenStoreQuery := `
	DROP TABLE IF EXISTS tokens CASCADE;
	CREATE TABLE tokens (
		token TEXT NOT NULL UNIQUE,
		creater_id INT NOT NULL,
		new_user_id INT,
		valid BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY(token),
		FOREIGN KEY(creater_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(new_user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	userMessageQuery := `
	DROP TABLE IF EXISTS users_messages CASCADE;
	CREATE TABLE users_messages (
		user_to_id INT,
		user_from_id INT NOT NULL,
		message_id INT NOT NULL,
		FOREIGN KEY(user_to_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(user_from_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
	);`

	channelMessageQuery := `
	DROP TABLE IF EXISTS channels_messages CASCADE;
	CREATE TABLE channels_messages (
		channel_id INT NOT NULL,
		message_id INT NOT NULL,
		FOREIGN KEY(channel_id) REFERENCES channels(id) ON DELETE CASCADE,
		FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
	);`

	queries := []string{
		userQuery, messageQuery, channelQuery, sidebarQuery, userChannelQuery,
		tokenStoreQuery, userMessageQuery, channelMessageQuery,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return errors.Wrapf(err, "Query: %v", query)
		}
	}

	return nil
}
