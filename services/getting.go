package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type getter struct {
	DB store.Getter
}

func NewGetter(db store.Getter) (sidebar.Getter, error) {
	return &getter{
		DB: db,
	}, nil
}

func (g *getter) GetUser(id int) (*sidebar.User, error) {
	return g.DB.GetUser(id)
}

func (g *getter) GetChannel(id int) (*sidebar.Channel, error) {
	return g.DB.GetChannel(id)
}

func (g *getter) GetMessage(id int) (*sidebar.WebSocketMessage, error) {
	return g.DB.GetMessage(id)
}

func (g *getter) GetUsers() ([]*sidebar.User, error) {
	return g.DB.GetUsers()
}

func (g *getter) GetChannels() ([]*sidebar.Channel, error) {
	return g.DB.GetChannels()
}

func (g *getter) GetMessages() ([]*sidebar.WebSocketMessage, error) {
	return g.DB.GetMessages()
}

func (g *getter) GetUsersInChannel(id int) ([]*sidebar.User, error) {
	return g.DB.GetUsersInChannel(id)
}

func (g *getter) GetChannelsForUser(id int) ([]*sidebar.Channel, error) {
	return g.DB.GetChannelsForUser(id)
}

func (g *getter) GetMessagesInChannel(id int) ([]*sidebar.WebSocketMessage, error) {
	return g.DB.GetMessagesInChannel(id)
}

func (g *getter) GetMessagesFromUser(id int) ([]*sidebar.WebSocketMessage, error) {
	return g.DB.GetMessagesFromUser(id)
}

func (g *getter) GetMessagesToUser(id int) ([]*sidebar.WebSocketMessage, error) {
	return g.DB.GetMessagesToUser(id)
}
