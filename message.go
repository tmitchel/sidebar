package sidebar

// event codes
const (
	EventMessage      = 1
	EventTyping       = 2
	EventStartSpinOff = 3
)

// ChatMessage represents a message sent over
// the Websocket connection.
type ChatMessage struct {
	ID       string `json:"id"`
	Event    int64  `json:"event"`
	Content  string `json:"content"`
	ToUser   string `json:"to_user"`
	FromUser string `json:"from_user"`
	Channel  string `json:"channel"`
}

// ChannelUpdate is sent over the Websocket connection
// to alert users of new channels, updates to an existing
// channel's information, etc.
type ChannelUpdate struct {
	Channel
	Message string `json:"message"`
}

// Alert is used to send news of an update to users.
type Alert struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

// WebsocketMessage represents anything that can
// be sent over the Websocket connection.
type WebsocketMessage struct {
	// Type of message so receiver can decide how to decode
	Type string

	// information being sent
	Payload interface{}
}
