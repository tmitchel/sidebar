package store

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"

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

// Database provides methods to query the database.
type Database interface {
	CreateUser(*sidebar.User) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
	CreateSpinoff(*sidebar.Spinoff) (*sidebar.Spinoff, error)

	GetUser(int) (*sidebar.User, error)
	GetChannel(int) (*sidebar.Channel, error)
	GetMessage(int) (*sidebar.WebSocketMessage, error)
	GetSpinoff(int) (*sidebar.Spinoff, error)

	GetUsers() ([]*sidebar.User, error)
	GetChannels() ([]*sidebar.Channel, error)
	GetSpinoffs() ([]*sidebar.Spinoff, error)
	GetMessages() ([]*sidebar.WebSocketMessage, error)

	GetUsersInChannel(int) ([]*sidebar.User, error)
	GetUsersInSpinoff(int) ([]*sidebar.User, error)

	GetMessagesInChannel(int) ([]*sidebar.WebSocketMessage, error)
	GetMessagesInSpinoff(int) ([]*sidebar.WebSocketMessage, error)
	GetMessagesFromUser(int) ([]*sidebar.WebSocketMessage, error)
	GetMessagesToUser(int) ([]*sidebar.WebSocketMessage, error)

	UserForAuth(string) (*sidebar.User, error)
	CheckToken(string) (*sidebar.User, error)

	Close()
}

type database struct {
	*sql.DB
}

// New connects to the postgres database, runs migrations,
// and returns that connection.
func New() (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, psqluser, password, dbname)
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

func (d *database) CreateUser(u *sidebar.User) (*sidebar.User, error) {
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

func (d *database) GetUser(id int) (*sidebar.User, error) {
	var u user
	row := psql.Select("id", "display_name", "email", "password").
		From("users").RunWith(d).QueryRow()
	err := row.Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}

	return u.ToModel(), nil
}

func (d *database) GetUsersInChannel(id int) ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "display_name", "email", "password").
		From("users").Join("user_channel uc ON ( uc.user_id = id )").
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

func (d *database) GetUsersInSpinoff(id int) ([]*sidebar.User, error) {
	var users []*sidebar.User
	rows, err := psql.Select("id", "display_name", "email", "password").
		From("users").Join("user_spinoff us ON ( us.user_id = id )").
		Where(sq.Eq{"us.spinoff_id": id}).RunWith(d).Query()
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

func (d *database) CheckToken(token string) (*sidebar.User, error) {
	var u user
	var t time.Time
	err := psql.Select("id", "display_name", "email", "password", "t.created_at").
		From("users").Join("tokens t ON (t.user_id = id)").
		Where(sq.Eq{"t.token": token}).
		RunWith(d).QueryRow().Scan(&u.ID, &u.DisplayName, &u.Email, &u.Password, &t)
	if err != nil || u.ID == 0 {
		return nil, errors.New("User doesn't have a token")
	}

	// this is how I will be able to tell if the token has expired. This will be
	// lifted into the service layer later.
	// if t.Add(30 * 24 * time.Hour).After(time.Now()) {
	// _, err := psql.Delete("tokens").Where(sq.Eq{"token": token}).RunWith(d).Exec()
	// if err != nil {
	// 	return nil, errors.New("Unable to delete token")
	// }
	// return nil, errors.New("Token has expired")
	// }

	return u.ToModel(), nil
}

func (d *database) CreateChannel(c *sidebar.Channel) (*sidebar.Channel, error) {
	var id int
	dchannel := channelFromModel(c)
	err := psql.Insert("channels").
		Columns("display_name").Values(dchannel.Name).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	dchannel.ID = id
	return dchannel.ToModel(), nil
}

func (d *database) GetChannel(id int) (*sidebar.Channel, error) {
	var c channel
	row := psql.Select("id", "display_name").From("channels").RunWith(d).QueryRow()
	err := row.Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, err
	}

	return c.ToModel(), nil
}

func (d *database) CreateSpinoff(s *sidebar.Spinoff) (*sidebar.Spinoff, error) {
	var id int
	dspinoff := spinoffFromModel(s)
	err := psql.Insert("spinoff").
		Columns("display_name", "parent_id").Values(dspinoff.Name, dspinoff.Parent).
		Suffix("RETURNING id").
		RunWith(d).QueryRow().Scan(&id)

	if err != nil {
		return nil, err
	}

	dspinoff.ID = id
	return dspinoff.ToModel(), nil
}

func (d *database) GetSpinoff(id int) (*sidebar.Spinoff, error) {
	var s spinoff
	row := psql.Select("id", "display_name", "parent_id").From("spinoffs").RunWith(d).QueryRow()
	err := row.Scan(&s.ID, &s.Name, &s.Parent)
	if err != nil {
		return nil, err
	}

	return s.ToModel(), nil
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

	dmessage.ID = id
	return dmessage.ToModel(), nil
}

func (d *database) GetMessage(id int) (*sidebar.WebSocketMessage, error) {
	var m webSocketMessage
	row := psql.Select("id", "event", "content").From("messages").RunWith(d).QueryRow()
	err := row.Scan(&m.ID, &m.Event, &m.Content)
	if err != nil {
		return nil, err
	}

	return m.ToModel(), nil
}

func (d *database) GetMessagesInChannel(id int) ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("channels_messages cm ON ( cm.message_id = id )").
		Where(sq.Eq{"cm.channel_id": id}).Where(sq.Eq{"cm.is_spinoff": false}).
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

func (d *database) GetMessagesInSpinoff(id int) ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "content").From("messages").
		Join("channels_messages cm ON ( cm.message_id = id )").
		Where(sq.Eq{"cm.channel_id": id}).Where(sq.Eq{"cm.is_spinoff": true}).
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
	var channels []*sidebar.Channel
	rows, err := psql.Select("id", "display_name").From("channels").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any channels")
	}

	for rows.Next() {
		var c channel
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, errors.New("Error scanning channels")
		}

		channels = append(channels, c.ToModel())
	}

	return channels, nil
}

func (d *database) GetSpinoffs() ([]*sidebar.Spinoff, error) {
	var spinoffs []*sidebar.Spinoff
	rows, err := psql.Select("id", "display_name", "parent").From("spinoffs").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any spinoffs")
	}

	for rows.Next() {
		var d spinoff
		err := rows.Scan(&d.ID, &d.Name, &d.Parent)
		if err != nil {
			return nil, errors.New("Error scanning spinoffs")
		}

		spinoffs = append(spinoffs, d.ToModel())
	}

	return spinoffs, nil
}

func (d *database) GetMessages() ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	rows, err := psql.Select("id", "event", "content").From("messages").
		RunWith(d).Query()
	if err != nil {
		return nil, errors.New("Unable to find any messages")
	}

	for rows.Next() {
		var w webSocketMessage
		err := rows.Scan(&w.ID, &w.Event, &w.Content)
		if err != nil {
			return nil, errors.New("Error scanning messages")
		}

		messages = append(messages, w.ToModel())
	}

	return messages, nil
}

func migrations(db *sql.DB) error {
	userQuery := `
	DROP TABLE IF EXISTS users CASCADE;
	CREATE TABLE users (
		id SERIAL,
		display_name TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		PRIMARY KEY(id)
	);`

	messageQuery := `
	DROP TABLE IF EXISTS messages CASCADE;
	CREATE TABLE messages (
		id SERIAL,
		content TEXT NOT NULL,
		event INT NOT NULL,
		PRIMARY KEY(id)
	);`

	channelQuery := `
	DROP TABLE IF EXISTS channels CASCADE;
	CREATE TABLE channels (
		id SERIAL,
		display_name TEXT UNIQUE NOT NULL,
		PRIMARY KEY(id)
	);`

	spinoffQuery := `
	DROP TABLE IF EXISTS spinoffs CASCADE;
	CREATE TABLE spinoffs (
		id SERIAL,
		display_name TEXT UNIQUE NOT NULL,
		parent_id INT NOT NULL,
		PRIMARY KEY(id),
		FOREIGN KEY (parent_ID) REFERENCES channels(id) ON DELETE CASCADE
	);`

	userChannelQuery := `
	DROP TABLE IF EXISTS user_channel CASCADE;
	CREATE TABLE user_channel (
		user_id INT REFERENCES users (id) ON UPDATE CASCADE,
		channel_id INT REFERENCES channels (id) ON UPDATE CASCADE,
		CONSTRAINT user_channel_pkey PRIMARY KEY (user_id, channel_id)
	);`

	userSpinoffQuery := `
	DROP TABLE IF EXISTS user_spinoff CASCADE;
	CREATE TABLE user_spinoff (
		user_id INT REFERENCES users (id) ON UPDATE CASCADE,
		spinoff_id INT REFERENCES spinoffs (id) ON UPDATE CASCADE,
		CONSTRAINT user_spinoff_pkey PRIMARY KEY (user_id, spinoff_id)
	);`

	tokenStoreQuery := `
	DROP TABLE IF EXISTS tokens CASCADE;
	CREATE TABLE tokens (
		token TEXT NOT NULL,
		user_id INT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY(token),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	userMessageQuery := `
	DROP TABLE IF EXISTS users_messages CASCADE;
	CREATE TABLE users_messages (
		user_to_id INT NOT NULL,
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
		is_spinoff BOOL DEFAULT FALSE,
		FOREIGN KEY(channel_id) REFERENCES channels(id) ON DELETE CASCADE,
		FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
	);`

	queries := []string{
		userQuery, messageQuery, channelQuery, spinoffQuery, userChannelQuery, userSpinoffQuery,
		tokenStoreQuery, userMessageQuery, channelMessageQuery,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}
