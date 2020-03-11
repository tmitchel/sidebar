package sidebar

// event codes
const (
	EventMessage      = 1
	EventTyping       = 2
	EventStartSpinOff = 3
)

// WebSocketMessage represents a message sent over
// the Websocket connection.
type WebSocketMessage struct {
	ID       int    `json:"id"`
	Event    int64  `json:"event"`
	Content  string `json:"content"`
	ToUser   int    `json:"to_user"`
	FromUser int    `json:"from_user"`
	Channel  int    `json:"channel"`
}
