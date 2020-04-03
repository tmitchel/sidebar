package services

import (
	"errors"

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

func (c *creater) NewToken(userID int) (string, error) {
	token := uuid.New().String()
	err := c.DB.NewToken(token, userID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *creater) CreateUser(u *sidebar.User, token string) (*sidebar.User, error) {
	hashed, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Error hashing password")
	}

	u.Password = hashed
	if u.ProfileImg == "" {
		u.ProfileImg = "https://randomuser.me/api/portraits/women/81.jpg"
	}
	return c.DB.CreateUser(u, token)
}

func (c *creater) CreateChannel(ch *sidebar.Channel) (*sidebar.Channel, error) {
	return c.DB.CreateChannel(ch)
}

func (c *creater) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	return c.DB.CreateMessage(m)
}
