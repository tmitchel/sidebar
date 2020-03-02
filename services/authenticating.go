package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type auth struct {
	DB store.Database
}

func NewAuthenticater(db store.Database) (sidebar.Authenticater, error) {
	return &auth{
		DB: db,
	}, nil
}

func (a *auth) Validate(email, password string) (*sidebar.User, error) {
	authUser, err := a.DB.UserForAuth(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(authUser.Password, []byte(password)); err != nil {
		return nil, errors.New("Incorrect password")
	}

	return a.DB.GetUser(authUser.ID)
}
