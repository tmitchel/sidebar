package mocks

import "github.com/tmitchel/sidebar"

type Creater struct {
	nUsers int
	Users  map[int]*sidebar.User
}

func NewCreater() Creater {
	return Creater{
		nUsers: 0,
		Users:  make(map[int]*sidebar.User),
	}
}

func (m *Creater) CreateUser(user *sidebar.User) (*sidebar.User, error) {
	user.ID = m.nUsers + 1
	m.nUsers++
	m.Users[user.ID] = user
	return user, nil
}

func (m *Creater) CreateChannel(*sidebar.Channel) (*sidebar.Channel, error) {
	return nil, nil
}

func (m *Creater) CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	return nil, nil
}
