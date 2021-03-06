package services

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type auth struct {
	DB store.Database
}

// NewAuthenticater wraps a database connection with an *auth that
// implements the sidebar.Authenticater interface.
func NewAuthenticater(db store.Database) (sidebar.Authenticater, error) {
	return &auth{
		DB: db,
	}, nil
}

// Validate gets the requested user from the database, checks the given password,
// then returns the full user if the password is correct.
func (a *auth) Validate(email, password, wid string) (*sidebar.User, error) {
	authUser, err := a.DB.UserForAuth(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(authUser.Password, []byte(password)); err != nil {
		return nil, errors.Wrap(err, "Incorrect password")
	}

	user, err := a.DB.GetUser(authUser.ID)
	if err != nil {
		return nil, err
	}

	if err := a.DB.UserInWorkspace(user.ID, wid); err != nil {
		return nil, err
	}

	return user, nil
}
