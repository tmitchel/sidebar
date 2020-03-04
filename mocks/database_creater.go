package mocks

import "github.com/tmitchel/sidebar"

type MockDatabaseCreater struct {
	Content map[string]interface{}
}

func (m *MockDatabaseCreater) CreateUser(*sidebar.User) (*sidebar.User, error) {
	return nil, nil
}

func (m *MockDatabaseCreater) CreateChannel(*sidebar.Channel) (*sidebar.Channel, error) {
	return nil, nil
}

func (m *MockDatabaseCreater) CreateMessage(*sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
	return nil, nil
}

func (m *MockDatabaseCreater) CreateSpinoff(*sidebar.Spinoff) (*sidebar.Spinoff, error) {
	return nil, nil
}
