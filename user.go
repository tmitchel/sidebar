package sidebar

// User represents a basic user of sidebar. They can be members
// of multiple channels, sidebars, etc.
type User struct {
	ID          int    `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    []byte `json:"-"`
	ProfileImg  string `json:"profile_image"`
}

type SignupUser struct {
	ID          int    `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password"`
	ProfileImg  string `json:"profile_image"`
}

type CompleteUser struct {
	User     User
	Channels []*ChannelForUser
}
