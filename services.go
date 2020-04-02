package sidebar

// Authenticater provides methods to check that a
// user has provided proper login information or
// a valid token.
type Authenticater interface {
	Validate(string, string) (*User, error)
}

// Creater provides methods to create new objects
// that will be persisted to disk.
type Creater interface {
	NewToken(int) (string, error)
	CreateUser(*User, string) (*User, error)
	CreateChannel(*Channel) (*Channel, error)
	CreateMessage(*WebSocketMessage) (*WebSocketMessage, error)
}

type Deleter interface {
	DeleteUser(int) (*User, error)
	DeleteChannel(int) (*Channel, error)
}

type Adder interface {
	AddUserToChannel(int, int) error
	RemoveUserFromChannel(int, int) error
}

type Getter interface {
	GetUser(int) (*User, error)
	GetChannel(int) (*Channel, error)
	GetMessage(int) (*WebSocketMessage, error)

	GetUsers() ([]*User, error)
	GetChannels() ([]*Channel, error)
	GetMessages() ([]*WebSocketMessage, error)

	GetUsersInChannel(int) ([]*User, error)
	GetChannelsForUser(int) ([]*Channel, error)

	GetMessagesInChannel(int) ([]*WebSocketMessage, error)
	GetMessagesFromUser(int) ([]*WebSocketMessage, error)
	GetMessagesToUser(int) ([]*WebSocketMessage, error)
}
