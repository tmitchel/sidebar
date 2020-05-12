package services

import (
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

type adder struct {
	DB store.Database
}

// NewAdder takes the database dependency and uses it to implement
// the sidebar.Adder interface. This interface is used to create new
// objects and store them in the database.
func NewAdder(db store.Database) (sidebar.Adder, error) {
	return &adder{
		DB: db,
	}, nil
}

// AddUserToChannel checks if the current user and channel are in the
// provided workspace. If so, the user is added to the channel.
func (a *adder) AddUserToChannel(userID, channelID, workID string) error {
	err := a.DB.UserInWorkspace(userID, workID)
	if err != nil {
		return err
	}

	err = a.DB.ChannelInWorkspace(channelID, workID)
	if err != nil {
		return err
	}

	return a.DB.AddUserToChannel(userID, channelID)
}

// RemoveUserFromChannel checks if the current user and channel are in the
// provided workspace. If so, the user is removed from the channel.
func (a *adder) RemoveUserFromChannel(userID, channelID, workID string) error {
	err := a.DB.UserInWorkspace(userID, workID)
	if err != nil {
		return err
	}

	err = a.DB.ChannelInWorkspace(channelID, workID)
	if err != nil {
		return err
	}

	return a.DB.RemoveUserFromChannel(userID, channelID)
}

// ResolveChannel checks if the channel is part of the current workspace.
// If so, the channel's "Resolved" state if flipped.
func (a *adder) ResolveChannel(id, workID string) error {
	err := a.DB.ChannelInWorkspace(id, workID)
	if err != nil {
		return err
	}
	return a.DB.ResolveChannel(id)
}

// AddUserToWorkspace confirms the user provided the correct token
// for joining the workspace and adds them on success.
func (a *adder) AddUserToWorkspace(uid, wid, token string) error {
	storedToken, err := a.DB.GetWorkspaceToken(wid)
	if err != nil {
		return err
	}

	if token != storedToken {
		return errors.Errorf("Token %v doesn't match %v", token, storedToken)
	}

	return a.DB.AddUserToWorkspace(uid, wid)
}
