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
	ID       int    `json:"-"`
	Event    int64  `json:"event"`
	Content  string `json:"content"`
	ToUser   int    `json:"-"`
	FromUser int    `json:"-"`
	Channel  int    `json:"-"`
}
