package sidebar

type Spinoff struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"display_name,omitempty"`
	Parent int    `json:"parent_id,omitempty"`
}
