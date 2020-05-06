package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/tmitchel/sidebar"
)

type CompleteUser struct {
	User     sidebar.User
	Channels []*ChannelForUser
}

type SignupUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	ProfileImg  string `json:"profile_image"`
}

type ChannelForUser struct {
	Channel sidebar.Channel
	Member  bool
}

type CompleteChannel struct {
	Channel           sidebar.Channel
	UsersInChannel    []*sidebar.User
	MessagesInChannel []*sidebar.WebSocketMessage
}

type updatePass struct {
	ID          string
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type Token struct {
	UserID        string
	Email         string
	UserName      string
	Authenticated bool
	jwt.StandardClaims
}

type auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
