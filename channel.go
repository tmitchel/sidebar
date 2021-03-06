package sidebar

// Channel contains a chat centered around a specific topic.
type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Details   string `json:"details"`
	Image     string `json:"display_image"`
	IsSidebar bool   `json:"is_sidebar"`
	Parent    string `json:"parent"`
	Direct    bool   `json:"direct"`
	Resolved  bool   `json:"resolved"`
}
