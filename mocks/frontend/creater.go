package frontend

import "github.com/tmitchel/sidebar"

type creater struct {
	Users    map[int]*sidebar.User
	Channels map[int]*sidebar.Channel
	Messages map[int]*sidebar.WebSocketMessage
}

func NewCreater() sidebar.Creater {
	return &creater{
		Users:    make(map[int]*sidebar.User),
		Channels: make(map[int]*sidebar.Channel),
		Messages: make(map[int]*sidebar.WebSocketMessage),
	}
}

func (c *creater) CreateUser(user *sidebar.User) (*sidebar.User, error) {
	user.ID = len(c.Users) + 10
	c.Users[user.ID] = user
	return user, nil
}

func (c *creater) CreateChannel(channel *sidebar.Channel) (*sidebar.Channel, error) {
	channel.ID = len(c.Channels) + 10
	c.Channels[channel.ID] = channel
	return channel, nil
}

func (c *creater) CreateMessage(message *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	message.ID = len(c.Messages) + 10
	c.Messages[message.ID] = message
	return message, nil
}
