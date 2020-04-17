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
	NewToken(string) (string, error)
	CreateUser(*User, string) (*User, error)
	CreateChannel(*Channel) (*Channel, error)
	CreateMessage(*WebSocketMessage) (*WebSocketMessage, error)
}

type Deleter interface {
	DeleteUser(string) (*User, error)
	DeleteChannel(string) (*Channel, error)
}

type Adder interface {
	ResolveChannel(string) error
	AddUserToChannel(string, string) error
	RemoveUserFromChannel(string, string) error
}

type Getter interface {
	GetUser(string) (*User, error)
	GetChannel(string) (*Channel, error)
	GetMessage(string) (*WebSocketMessage, error)

	GetUsers() ([]*User, error)
	GetChannels() ([]*Channel, error)
	GetMessages() ([]*WebSocketMessage, error)

	GetUsersInChannel(string) ([]*User, error)
	GetChannelsForUser(string) ([]*Channel, error)

	GetMessagesInChannel(string) ([]*WebSocketMessage, error)
	GetMessagesFromUser(string) ([]*WebSocketMessage, error)
	GetMessagesToUser(string) ([]*WebSocketMessage, error)
}

type Updater interface {
	UpdateUserInfo(*User) error
	UpdateUserPassword(string, []byte, []byte) error
}
