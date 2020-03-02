package sidebar

type User struct {
	ID          int    `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"-"`
}
