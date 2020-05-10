package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/tmitchel/sidebar"
)

// SignupUser is used to decode JSON sent from the frontend
// for creating a new account.
type SignupUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	ProfileImg  string `json:"profile_image"`
}

// JWTToken contains information to be stored in a JWT
// on the client side.
type JWTToken struct {
	UserID        string
	Authenticated bool
	jwt.StandardClaims
}

// UserWithChannels provides the user's information along
// with all the channels to which the user belongs.
type UserWithChannels struct {
	User     sidebar.User
	Channels []*ChannelWithMemberInfo
}

// ChannelWithMemberInfo contains a channel, a user id, and a bool
// telling whether the user is a member of this channel. This is
// stored within a UserWithChannels struct.
type ChannelWithMemberInfo struct {
	sidebar.Channel
	MemberID string
	Member   bool
}

// ChannelWithUsersAndMessages provides a channel along with
// information about all users and messages in the channel.
type ChannelWithUsersAndMessages struct {
	Channel           sidebar.Channel
	UsersInChannel    []*sidebar.User
	MessagesInChannel []*sidebar.ChatMessage
}

// PasswordUpdate is used to decode requests to update the
// user's password.
type PasswordUpdate struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// AuthInfo is used to decode requests to log in.
type AuthInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
