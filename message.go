package sidebar

// event codes
const (
	EventMessage      = 1
	EventStartSpinOff = 2
)

// WebSocketMessage represents a message sent over
// the Websocket connection.
type WebSocketMessage struct {
	ID      int    `json:"-"`
	Event   int64  `json:"event"`
	Content string `json:"content"`
}
