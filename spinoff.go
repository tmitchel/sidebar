package sidebar

// Spinoff represents a channel created by a spinoff
// command. This means the channel is created as a
// response to a message in another channel (the parent).
type Spinoff struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"display_name,omitempty"`
	Parent int    `json:"parent_id,omitempty"`
}
