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
	CreateUser(*User) (*User, error)
	CreateChannel(*Channel) (*Channel, error)
	CreateMessage(*WebSocketMessage) (*WebSocketMessage, error)
}
