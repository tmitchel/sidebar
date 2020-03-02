package sidebar

type Authenticater interface {
	Validate(string, string) (*User, error)
}

type Creater interface {
	CreateUser(*User) (*User, error)
	CreateChannel(*Channel) (*Channel, error)
	CreateSpinoff(*Spinoff) (*Spinoff, error)
	CreateMessage(*WebSocketMessage) (*WebSocketMessage, error)
}
