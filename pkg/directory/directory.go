package directory

// A User represents a user in a directory.
type User struct {
	ID          string   `json:"id,omitempty"`
	GroupIDs    []string `json:"group_ids,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	Email       string   `json:"email,omitempty"`
}

// A Group represents a group in a directory.
type Group struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
