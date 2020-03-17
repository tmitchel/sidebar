package frontend

import "github.com/tmitchel/sidebar"

type adder struct {
	UserToChannel map[int]int
}

func NewAdder() sidebar.Adder {
	return &adder{
		UserToChannel: make(map[int]int),
	}
}

func (a *adder) AddUserToChannel(userID, channelID int) error {
	a.UserToChannel[userID] = channelID
	return nil
}
