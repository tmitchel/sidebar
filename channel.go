package sidebar

// Channel contains a chat centered around a specific topic.
type Channel struct {
	ID        int
	Name      string
	IsSidebar bool
	Parent    int
	Direct    bool
	Resolved  bool
}

type CompleteChannel struct {
	Channel           Channel
	UsersInChannel    []*User
	MessagesInChannel []*WebSocketMessage
}
