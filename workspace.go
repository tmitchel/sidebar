package sidebar

// Workspace represents a set of channels and users in
// one group.
type Workspace struct {
	ID          string `json:"id"`
	Token       string `json:"token"`
	DisplayName string `json:"display_name"`
	DisplayImg  string `json:"display_image"`
}
