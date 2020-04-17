package store

import (
	"github.com/tmitchel/sidebar"
)

type user struct {
	ID          string
	DisplayName string
	Email       string
	Password    []byte
	ProfileImg  string
}

// userFromModel converts the normal sidebar.User model
// into a user which has properties only useful for the
// database.
func userFromModel(u *sidebar.User) *user {
	return &user{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Password:    u.Password,
		ProfileImg:  u.ProfileImg,
	}
}

func (u *user) ToModel() *sidebar.User {
	return &sidebar.User{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Password:    u.Password,
		ProfileImg:  u.ProfileImg,
	}
}

type userForAuth struct {
	ID       string
	Password []byte
}

func (u *userForAuth) ToModel() *sidebar.User {
	return &sidebar.User{
		ID:          u.ID,
		DisplayName: "-",
		Email:       "-",
		Password:    u.Password,
	}
}
