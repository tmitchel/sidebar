package sidebar

// Authenticater provides methods to check that a
// user has provided proper login information or
// a valid token.
type Authenticater interface {
	Validate(string, string, string) (*User, error)
}

// Creater provides methods to create new objects
// that will be persisted to disk.
type Creater interface {
	CreateWorkspace(*Workspace) (*Workspace, error)
	CreateUser(*User) (*User, error)
	CreateChannel(*Channel, string) (*Channel, error)
	CreateMessage(*ChatMessage) (*ChatMessage, error)
}

type Deleter interface {
	DeleteUser(string) (*User, error)
	DeleteChannel(string) (*Channel, error)
}

type Adder interface {
	ResolveChannel(string, string) error
	AddUserToChannel(string, string, string) error
	RemoveUserFromChannel(string, string, string) error
}

type Getter interface {
	GetWorkspacesForUser(string) ([]*Workspace, error)
	GetWorkspaces() ([]*Workspace, error)

	GetUser(string) (*User, error)
	GetChannel(string) (*Channel, error)
	GetMessage(string) (*ChatMessage, error)

	GetUsers(string) ([]*User, error)
	GetChannels(string) ([]*Channel, error)
	GetMessages(string) ([]*ChatMessage, error)

	GetUsersInChannel(string, string) ([]*User, error)
	GetChannelsForUser(string, string) ([]*Channel, error)

	GetMessagesInChannel(string, string) ([]*ChatMessage, error)
	GetMessagesFromUser(string, string) ([]*ChatMessage, error)
	GetMessagesToUser(string, string) ([]*ChatMessage, error)
}

type Updater interface {
	UpdateUserInfo(*User) error
	UpdateChannelInfo(*Channel) error
	UpdateUserPassword(string, []byte, []byte) error
}
