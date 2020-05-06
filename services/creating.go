package services

import (
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

type creater struct {
	DB store.Creater
}

// NewCreater wraps a database connection with a *creater that
// implements the sidebar.Creater interface.
func NewCreater(db store.Creater) (sidebar.Creater, error) {
	return &creater{
		DB: db,
	}, nil
}

// NewToken creates a new token for inviting new users. The token
// is stored in the database and is checked when the new user
// tries to make an account.
func (c *creater) NewToken(userID string) (string, error) {
	token := uuid.New().String()
	err := c.DB.NewToken(token, userID)
	if err != nil {
		return "", err
	}
	return token, nil
}

// CreateUser takes the new user's information and the token they were sent
// to create an account. The user is assigned an id, their password is hashed
// and they are given a default profile image if they didn't provide one. The
// user is stored in the database.
func (c *creater) CreateUser(u *sidebar.User, token string) (*sidebar.User, error) {
	u.ID = uuid.New().String()

	hashed, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "Error hashing password")
	}

	u.Password = hashed
	if u.ProfileImg == "" {
		u.ProfileImg = "https://randomuser.me/api/portraits/women/81.jpg"
	}
	return c.DB.CreateUser(u, token)
}

// CreateChannel takes the information sent for creating a new channel,
// gives it an id and a default image if one isn't provided. The channel is
// saved.
func (c *creater) CreateChannel(ch *sidebar.Channel) (*sidebar.Channel, error) {
	ch.ID = uuid.New().String()
	if ch.Image == "" {
		ch.Image = "https://randomuser.me/api/portraits/women/81.jpg"
	}
	return c.DB.CreateChannel(ch)
}

// CreateMessage gives the message an id and stores in the database.
func (c *creater) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	m.ID = uuid.New().String()
	return c.DB.CreateMessage(m)
}
