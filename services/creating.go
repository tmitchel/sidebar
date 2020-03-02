package services

import (
	"errors"

	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

type Creater interface {
	CreateUser(*sidebar.User) (*sidebar.User, error)
	CreateChannel(*sidebar.Channel) (*sidebar.Channel, error)
	CreateSpinoff(*sidebar.Spinoff) (*sidebar.Spinoff, error)
	CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error)
}

type creater struct {
	DB store.Database
}

func NewCreater(db store.Database) (Creater, error) {
	return &creater{
		DB: db,
	}, nil
}

func (c *creater) CreateUser(u *sidebar.User) (*sidebar.User, error) {
	hashed, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Error hashing password")
	}

	u.Password = hashed
	return c.DB.CreateUser(u)
}

func (c *creater) CreateChannel(ch *sidebar.Channel) (*sidebar.Channel, error) {
	return c.DB.CreateChannel(ch)
}

func (c *creater) CreateSpinoff(s *sidebar.Spinoff) (*sidebar.Spinoff, error) {
	return c.DB.CreateSpinoff(s)
}

func (c *creater) CreateMessage(m *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	return c.DB.CreateMessage(m)
}
