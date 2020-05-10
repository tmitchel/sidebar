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
