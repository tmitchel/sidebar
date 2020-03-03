package store

import (
	"github.com/tmitchel/sidebar"
)

type user struct {
	ID          int    `db:"id"`
	DisplayName string `db:"display_name"`
	Email       string `db:"email"`
	Password    []byte `db:"password"`
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
	}
}

func (u *user) ToModel() *sidebar.User {
	return &sidebar.User{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Password:    u.Password,
	}
}

type userForAuth struct {
	ID       int    `db:"id"`
	Password []byte `db:"password"`
}

func (u *userForAuth) ToModel() *sidebar.User {
	return &sidebar.User{
		ID:          u.ID,
		DisplayName: "-",
		Email:       "-",
		Password:    u.Password,
	}
}
