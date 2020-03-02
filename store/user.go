package store

import (
	"github.com/tmitchel/sidebar"
)

type user struct {
	ID          int    `db:"id"`
	DisplayName string `db:"display_name"`
	Email       string `db:"email"`
	Password    string `db:"password"`
}

func UserFromModel(u *sidebar.User) *user {
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
