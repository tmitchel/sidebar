package frontend

import (
	"errors"

	"github.com/tmitchel/sidebar"
)

type authenticater struct {
	Users map[int]*sidebar.User
}

func NewAuthenticater() sidebar.Authenticater {
	return &authenticater{
		Users: map[int]*sidebar.User{
			1: &sidebar.User{
				ID:          1,
				DisplayName: "user one",
				Email:       "userone@email.com",
				Password:    []byte("password"),
			},
			2: &sidebar.User{
				ID:          2,
				DisplayName: "user twp",
				Email:       "usertwp@email.com",
				Password:    []byte("password"),
			},
			3: &sidebar.User{
				ID:          3,
				DisplayName: "user three",
				Email:       "userthree@email.com",
				Password:    []byte("password"),
			},
		},
	}
}

func (a *authenticater) Validate(username, password string) (*sidebar.User, error) {
	var foundUser *sidebar.User
	for _, user := range a.Users {
		if user.DisplayName == username {
			foundUser = user
		}
	}

	if foundUser == nil {
		return nil, errors.New("Unable to locate user")
	}

	if string(foundUser.Password) != password {
		return nil, errors.New("Passwords do not match")
	}

	return foundUser, nil
}
