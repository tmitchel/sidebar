package services

import (
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type getter struct {
	DB store.Getter
}

// NewGetter takes the database dependency and uses it to implement
// the sidebar.Getter interface. This interface is used to retrieve
// objects from store the database.
func NewGetter(db store.Getter) (sidebar.Getter, error) {
	return &getter{
		DB: db,
	}, nil
}

func (g *getter) GetUser(id string) (*sidebar.User, error) {
	return g.DB.GetUser(id)
}

func (g *getter) GetChannel(id string) (*sidebar.Channel, error) {
	return g.DB.GetChannel(id)
}

func (g *getter) GetMessage(id string) (*sidebar.ChatMessage, error) {
	return g.DB.GetMessage(id)
}

func (g *getter) GetUsers() ([]*sidebar.User, error) {
	return g.DB.GetUsers()
}

func (g *getter) GetChannels() ([]*sidebar.Channel, error) {
	return g.DB.GetChannels()
}

func (g *getter) GetMessages() ([]*sidebar.ChatMessage, error) {
	return g.DB.GetMessages()
}

func (g *getter) GetUsersInChannel(id string) ([]*sidebar.User, error) {
	return g.DB.GetUsersInChannel(id)
}

func (g *getter) GetChannelsForUser(id string) ([]*sidebar.Channel, error) {
	return g.DB.GetChannelsForUser(id)
}

func (g *getter) GetMessagesInChannel(id string) ([]*sidebar.ChatMessage, error) {
	return g.DB.GetMessagesInChannel(id)
}

func (g *getter) GetMessagesFromUser(id string) ([]*sidebar.ChatMessage, error) {
	return g.DB.GetMessagesFromUser(id)
}

func (g *getter) GetMessagesToUser(id string) ([]*sidebar.ChatMessage, error) {
	return g.DB.GetMessagesToUser(id)
}
