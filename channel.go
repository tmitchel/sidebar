package sidebar

// Channel contains a chat centered around a specific topic.
type Channel struct {
	ID        int
	Name      string
	IsSidebar bool
	Parent    int
}

type CompleteChannel struct {
	Channel           Channel
	UsersInChannel    []*User
	MessagesInChannel []*WebSocketMessage
}
