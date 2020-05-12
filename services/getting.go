package services

import (
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type getter struct {
	DB store.Database
}

// NewGetter takes the database dependency and uses it to implement
// the sidebar.Getter interface. This interface is used to retrieve
// objects from store the database.
func NewGetter(db store.Database) (sidebar.Getter, error) {
	return &getter{
		DB: db,
	}, nil
}

func (g *getter) GetWorkspacesForUser(uid string) ([]*sidebar.Workspace, error) {
	return g.DB.GetWorkspacesForUser(uid)
}

func (g *getter) GetWorkspaces() ([]*sidebar.Workspace, error) {
	return g.DB.GetWorkspaces()
}

// GetUser returns the user with the given id.
func (g *getter) GetUser(id string) (*sidebar.User, error) {
	return g.DB.GetUser(id)
}

// GetChannel returns the channel with the given id.
func (g *getter) GetChannel(id string) (*sidebar.Channel, error) {
	return g.DB.GetChannel(id)
}

// GetMessage returns the message with the given id.
func (g *getter) GetMessage(id string) (*sidebar.ChatMessage, error) {
	return g.DB.GetMessage(id)
}

// GetUsers returns all users in the given workspace.
func (g *getter) GetUsers(wid string) ([]*sidebar.User, error) {
	users, err := g.DB.GetUsers()
	if err != nil {
		return nil, err
	}

	var usersInWS []*sidebar.User
	for _, u := range users {
		if err := g.DB.UserInWorkspace(u.ID, wid); err != nil {
			continue
		}
		usersInWS = append(usersInWS, u)
	}
	return usersInWS, nil
}

// GetChannels returns all channels in the given workspace.
func (g *getter) GetChannels(wid string) ([]*sidebar.Channel, error) {
	channels, err := g.DB.GetChannels()
	if err != nil {
		return nil, err
	}

	var channelsInWS []*sidebar.Channel
	for _, c := range channels {
		if err := g.DB.ChannelInWorkspace(c.ID, wid); err != nil {
			continue
		}
		channelsInWS = append(channelsInWS, c)
	}
	return channelsInWS, nil
}

// GetMessages returns all messages sent in a channel that is part of
// the current workspace.
func (g *getter) GetMessages(wid string) ([]*sidebar.ChatMessage, error) {
	messages, err := g.DB.GetMessages()
	if err != nil {
		return nil, err
	}

	var messagesInWS []*sidebar.ChatMessage
	for _, m := range messages {
		if g.DB.ChannelInWorkspace(m.Channel, wid); err != nil {
			continue
		}
		messagesInWS = append(messagesInWS, m)
	}

	return messagesInWS, nil
}

// GetUsersInChannel checks whether the given channel is in the current
// workspace. Users can only be added to channels in a workspace in which
// they are a member, so we don't need to check each individual user.
func (g *getter) GetUsersInChannel(id, wid string) ([]*sidebar.User, error) {
	if err := g.DB.ChannelInWorkspace(id, wid); err != nil {
		return nil, errors.Wrap(err, "Error getting channels in workspace")
	}

	return g.DB.GetUsersInChannel(id)
}

// GetChannelsForUser gets all the channels in which the user is a member.
// Users can join channels in any of their workspaces, so we need to check
// if each channel is part of the current workspace.
func (g *getter) GetChannelsForUser(id, wid string) ([]*sidebar.Channel, error) {
	channels, err := g.DB.GetChannelsForUser(id)
	if err != nil {
		return nil, err
	}

	var channelsInWS []*sidebar.Channel
	for _, c := range channels {
		if err := g.DB.ChannelInWorkspace(c.ID, wid); err != nil {
			continue
		}
		channelsInWS = append(channelsInWS, c)
	}
	return channelsInWS, nil
}

// GetMessagesInChannel returns all messages for the given channel after checking
// that the provided channel is in the current workspace.
func (g *getter) GetMessagesInChannel(id, wid string) ([]*sidebar.ChatMessage, error) {
	if err := g.DB.ChannelInWorkspace(id, wid); err != nil {
		return nil, err
	}
	return g.DB.GetMessagesInChannel(id)
}

// GetMessagesFromUser returns all messages sent by the user in a channel that is
// a member of the current workspace.
func (g *getter) GetMessagesFromUser(id, wid string) ([]*sidebar.ChatMessage, error) {
	messages, err := g.DB.GetMessagesFromUser(id)
	if err != nil {
		return nil, err
	}

	var messagesInWS []*sidebar.ChatMessage
	for _, m := range messages {
		if g.DB.ChannelInWorkspace(m.Channel, wid); err != nil {
			continue
		}
		messagesInWS = append(messagesInWS, m)
	}

	return messagesInWS, nil
}

// GetMessagesToUser returns all messages sent to the user in a channel that is
// a member of the current workspace.
func (g *getter) GetMessagesToUser(id, wid string) ([]*sidebar.ChatMessage, error) {
	messages, err := g.DB.GetMessagesToUser(id)
	if err != nil {
		return nil, err
	}

	var messagesInWS []*sidebar.ChatMessage
	for _, m := range messages {
		if g.DB.ChannelInWorkspace(m.Channel, wid); err != nil {
			continue
		}
		messagesInWS = append(messagesInWS, m)
	}

	return messagesInWS, nil
}
