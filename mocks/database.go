package mocks

// import (
// 	"database/sql"
// 	"errors"

// 	"github.com/tmitchel/sidebar"
// 	"github.com/tmitchel/sidebar/store"
// )

// type Database struct {
// 	Users         map[int]*sidebar.User
// 	Channels      map[int]*sidebar.Channel
// 	Messages      map[int]*sidebar.WebSocketMessage
// 	UserToChannel map[int][]int
// }

// func NewDatabase() store.Database {
// 	return &Database{
// 		Users: map[int]*sidebar.User{
// 			1: &sidebar.User{
// 				ID:          1,
// 				DisplayName: "user one",
// 				Email:       "userone@email.com",
// 				Password:    []byte("password"),
// 			},
// 			2: &sidebar.User{
// 				ID:          2,
// 				DisplayName: "user two",
// 				Email:       "usertwo@email.com",
// 				Password:    []byte("password"),
// 			},
// 			3: &sidebar.User{
// 				ID:          3,
// 				DisplayName: "user three",
// 				Email:       "userthree@email.com",
// 				Password:    []byte("password"),
// 			},
// 		},
// 		Channels: map[int]*sidebar.Channel{
// 			1: &sidebar.Channel{
// 				ID:        1,
// 				Name:      "channel one",
// 				IsSidebar: false,
// 			},
// 			2: &sidebar.Channel{
// 				ID:        2,
// 				Name:      "channel two",
// 				IsSidebar: true,
// 				Parent:    1,
// 			},
// 		},
// 		Messages: map[int]*sidebar.WebSocketMessage{
// 			1: &sidebar.WebSocketMessage{
// 				ID:       1,
// 				Event:    1,
// 				Content:  "message one",
// 				ToUser:   2,
// 				FromUser: 1,
// 				Channel:  1,
// 			},
// 			2: &sidebar.WebSocketMessage{
// 				ID:       2,
// 				Event:    1,
// 				Content:  "message two",
// 				ToUser:   0,
// 				FromUser: 2,
// 				Channel:  2,
// 			},
// 			3: &sidebar.WebSocketMessage{
// 				ID:       3,
// 				Event:    2,
// 				Content:  "",
// 				ToUser:   0,
// 				FromUser: 1,
// 				Channel:  1,
// 			},
// 		},
// 		UserToChannel: map[int][]int{
// 			1: []int{1, 2},
// 			2: []int{1, 2},
// 			3: []int{1},
// 		},
// 	}
// }

// func (d *Database) AddUserToChannel(userID, channelID int) error {
// 	d.UserToChannel[userID] = append(d.UserToChannel[userID], channelID)
// 	return nil
// }

// func (d *Database) Close() {

// }

// func (d *Database) CreateUser(user *sidebar.User) (*sidebar.User, error) {
// 	user.ID = len(d.Users) + 1
// 	d.Users[user.ID] = user
// 	return user, nil
// }

// func (d *Database) CreateChannel(channel *sidebar.Channel) (*sidebar.Channel, error) {
// 	channel.ID = len(d.Channels) + 1
// 	d.Channels[channel.ID] = channel
// 	return channel, nil
// }

// func (d *Database) CreateMessage(message *sidebar.WebSocketMessage) (*sidebar.WebSocketMessage, error) {
// 	message.ID = len(d.Messages) + 1
// 	d.Messages[message.ID] = message
// 	return message, nil
// }

// func (d *Database) DeleteChannel(id int) (*sidebar.Channel, error) {
// 	return nil, nil
// }

// func (d *Database) DeleteUser(id int) (*sidebar.User, error) {
// 	return nil, nil
// }

// func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
// 	return nil, nil
// }

// func (d *Database) GetUser(id int) (*sidebar.User, error) {
// 	user, ok := d.Users[id]
// 	if !ok {
// 		return nil, errors.New("doesn't exist")
// 	}
// 	return user, nil
// }

// func (d *Database) GetChannel(id int) (*sidebar.Channel, error) {
// 	channel, ok := d.Channels[id]
// 	if !ok {
// 		return nil, errors.New("doesn't exist")
// 	}
// 	return channel, nil
// }

// func (d *Database) GetMessage(id int) (*sidebar.WebSocketMessage, error) {
// 	message, ok := d.Messages[id]
// 	if !ok {
// 		return nil, errors.New("doesn't exist")
// 	}
// 	return message, nil
// }

// func (d *Database) GetUsers() ([]*sidebar.User, error) {
// 	var users []*sidebar.User
// 	for _, user := range d.Users {
// 		users = append(users, user)
// 	}
// 	return users, nil
// }

// func (d *Database) GetChannels() ([]*sidebar.Channel, error) {
// 	var channels []*sidebar.Channel
// 	for _, channel := range d.Channels {
// 		channels = append(channels, channel)
// 	}
// 	return channels, nil
// }

// func (d *Database) GetMessages() ([]*sidebar.WebSocketMessage, error) {
// 	var messages []*sidebar.WebSocketMessage
// 	for _, message := range d.Messages {
// 		messages = append(messages, message)
// 	}
// 	return messages, nil
// }

// func (d *Database) GetUsersInChannel(id int) ([]*sidebar.User, error) {
// 	channels := d.UserToChannel[id]
// }

// func (d *Database) GetChannelsForUser(id int) ([]*sidebar.Channel, error) {
// 	var channels []*sidebar.Channel
// 	channelIDs := d.UserToChannel[id]
// 	for _, c := range d.Channels {
// 		for _, id := range channelIDs {
// 			if c.ID == id {
// 				channels = append(channels, c)
// 			}
// 		}
// 	}
// 	return channels, nil
// }

// func (d *Database) GetMessagesInChannel(id int) ([]*sidebar.WebSocketMessage, error) {
// 	if id == 1 {
// 		return []*sidebar.WebSocketMessage{d.Messages[1], d.Messages[3]}, nil
// 	} else if id == 2 {
// 		return []*sidebar.WebSocketMessage{d.Messages[2]}, nil
// 	}

// 	return nil, nil
// }

// func (d *Database) GetMessagesFromUser(id int) ([]*sidebar.WebSocketMessage, error) {
// 	if id == 1 {
// 		return []*sidebar.WebSocketMessage{d.Messages[1], d.Messages[3]}, nil
// 	} else if id == 2 {
// 		return []*sidebar.WebSocketMessage{d.Messages[2]}, nil
// 	}
// 	return nil, nil
// }

// func (d *Database) GetMessagesToUser(id int) ([]*sidebar.WebSocketMessage, error) {
// 	if id == 1 {
// 		return []*sidebar.WebSocketMessage{d.Messages[2]}, nil
// 	}
// 	return nil, nil
// }
