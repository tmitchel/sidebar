package sidebar

// User represents a basic user of sidebar. They can be members
// of multiple channels, sidebars, etc.
type User struct {
	ID          int    `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    []byte `json:"-"`
}
