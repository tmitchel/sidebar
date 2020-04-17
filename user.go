package sidebar

// User represents a basic user of sidebar. They can be members
// of multiple channels, sidebars, etc.
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    []byte `json:"-"`
	ProfileImg  string `json:"profile_image"`
}

type SignupUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	ProfileImg  string `json:"profile_image"`
}

type CompleteUser struct {
	User     User
	Channels []*ChannelForUser
}
