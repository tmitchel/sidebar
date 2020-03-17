package frontend

import (
	"errors"

	"github.com/tmitchel/sidebar"
)

type getter struct {
	Users    map[int]*sidebar.User
	Channels map[int]*sidebar.Channel
	Messages map[int]*sidebar.WebSocketMessage
}

func NewGetter() sidebar.Getter {
	return &getter{
		Users: map[int]*sidebar.User{
			1: &sidebar.User{
				ID:          1,
				DisplayName: "user one",
				Email:       "userone@email.com",
				Password:    []byte("password"),
			},
			2: &sidebar.User{
				ID:          2,
				DisplayName: "user two",
				Email:       "usertwo@email.com",
				Password:    []byte("password"),
			},
			3: &sidebar.User{
				ID:          3,
				DisplayName: "user three",
				Email:       "userthree@email.com",
				Password:    []byte("password"),
			},
		},
		Channels: map[int]*sidebar.Channel{
			1: &sidebar.Channel{
				ID:        1,
				Name:      "channel one",
				IsSidebar: false,
			},
			2: &sidebar.Channel{
				ID:        2,
				Name:      "channel two",
				IsSidebar: true,
				Parent:    1,
			},
		},
		Messages: map[int]*sidebar.WebSocketMessage{
			1: &sidebar.WebSocketMessage{
				ID:       1,
				Event:    1,
				Content:  "message one",
				ToUser:   2,
				FromUser: 1,
				Channel:  1,
			},
			2: &sidebar.WebSocketMessage{
				ID:       2,
				Event:    1,
				Content:  "message two",
				ToUser:   0,
				FromUser: 2,
				Channel:  2,
			},
			3: &sidebar.WebSocketMessage{
				ID:       3,
				Event:    2,
				Content:  "",
				ToUser:   0,
				FromUser: 1,
				Channel:  1,
			},
		},
	}
}

func (g *getter) GetUser(id int) (*sidebar.User, error) {
	user, ok := g.Users[id]
	if !ok {
		return nil, errors.New("doesn't exist")
	}
	return user, nil
}

func (g *getter) GetChannel(id int) (*sidebar.Channel, error) {
	channel, ok := g.Channels[id]
	if !ok {
		return nil, errors.New("doesn't exist")
	}
	return channel, nil
}

func (g *getter) GetMessage(id int) (*sidebar.WebSocketMessage, error) {
	message, ok := g.Messages[id]
	if !ok {
		return nil, errors.New("doesn't exist")
	}
	return message, nil
}

func (g *getter) GetUsers() ([]*sidebar.User, error) {
	var users []*sidebar.User
	for _, user := range g.Users {
		users = append(users, user)
	}
	return users, nil
}

func (g *getter) GetChannels() ([]*sidebar.Channel, error) {
	var channels []*sidebar.Channel
	for _, channel := range g.Channels {
		channels = append(channels, channel)
	}
	return channels, nil
}

func (g *getter) GetMessages() ([]*sidebar.WebSocketMessage, error) {
	var messages []*sidebar.WebSocketMessage
	for _, message := range g.Messages {
		messages = append(messages, message)
	}
	return messages, nil
}

func (g *getter) GetUsersInChannel(id int) ([]*sidebar.User, error) {
	return []*sidebar.User{g.Users[1], g.Users[2]}, nil
}

func (g *getter) GetChannelsForUser(id int) ([]*sidebar.Channel, error) {
	if id == 1 || id == 2 {
		var channels []*sidebar.Channel
		for _, channel := range g.Channels {
			channels = append(channels, channel)
		}
		return channels, nil
	}

	return nil, nil
}

func (g *getter) GetMessagesInChannel(id int) ([]*sidebar.WebSocketMessage, error) {
	if id == 1 {
		return []*sidebar.WebSocketMessage{g.Messages[1], g.Messages[3]}, nil
	} else if id == 2 {
		return []*sidebar.WebSocketMessage{g.Messages[2]}, nil
	}

	return nil, nil
}

func (g *getter) GetMessagesFromUser(id int) ([]*sidebar.WebSocketMessage, error) {
	if id == 1 {
		return []*sidebar.WebSocketMessage{g.Messages[1], g.Messages[3]}, nil
	} else if id == 2 {
		return []*sidebar.WebSocketMessage{g.Messages[2]}, nil
	}
	return nil, nil
}

func (g *getter) GetMessagesToUser(id int) ([]*sidebar.WebSocketMessage, error) {
	if id == 1 {
		return []*sidebar.WebSocketMessage{g.Messages[2]}, nil
	}
	return nil, nil
}
