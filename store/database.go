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

type Database interface {
	CreateUser(*sidebar.User) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
	CreateSpinoff(*sidebar.Spinoff) (*sidebar.Spinoff, error)

	GetUser(int) (*sidebar.User, error)
	GetChannel(int) (*sidebar.Channel, error)
	GetMessage(int) (*sidebar.WebSocketMessage, error)
	GetSpinoff(int) (*sidebar.Spinoff, error)

	GetUsersInChannel(int) ([]*sidebar.User, error)
	GetUsersInSpinoff(int) ([]*sidebar.User, error)

	UserForAuth(string) (*sidebar.User, error)
	CheckToken(string) (*sidebar.User, error)

	Close()
}

type database struct {
	*sql.DB
}

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

func (d *database) Close() {
	d.DB.Close()
}

func (d *database) CreateUser(u *sidebar.User) (*sidebar.User, error) {
	var id int
	duser := UserFromModel(u)
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
	dchannel := ChannelFromModel(c)
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
	dspinoff := SpinoffFromModel(s)
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
	dmessage := MessageFromModel(m)
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
		user_id INT REFERENCES users (user_id) ON UPDATE CASCADE,
		channel_id INT REFERENCES channels (channel_id) ON UPDATE CASCADE,
		CONSTRAINT user_channel_pkey PRIMARY KEY (user_id, channel_id)
	);`

	userSpinoffQuery := `
	DROP TABLE IF EXISTS user_spinoff CASCADE;
	CREATE TABLE user_spinoff (
		user_id INT REFERENCES users (user_id) ON UPDATE CASCADE,
		spinoff_id INT REFERENCES spinoffs (spinoff_id) ON UPDATE CASCADE,
		CONSTRAINT user_spinoff_pkey PRIMARY KEY (user_id, spinoff_id)
	);`

	tokenStoreQuery := `
	DROP TABLE IF EXISTS tokens CASCADE;	
	CREATE TABLE tokens (
		token TEXT NOT NULL,
		user_id INT NOT NULL,
		created_at TIMESTAMPZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY(token),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	queries := []string{
		userQuery, messageQuery, channelQuery, spinoffQuery, userChannelQuery, userSpinoffQuery,
		tokenStoreQuery,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}
