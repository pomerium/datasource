package directory

const (
	GroupRecordType = "pomerium.io/DirectoryGroup"
	StateRecordType = "pomerium.io/DirectoryState"
	UserRecordType  = "pomerium.io/DirectoryUser"
)

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

// A Bundle is a bundle of directory data.
type Bundle map[string]any

// NewBundle creates a new Bundle of directory data.
func NewBundle(groups []Group, users []User, state map[string]any) Bundle {
	// don't serve null, but an empty array instead
	if groups == nil {
		groups = make([]Group, 0)
	}
	if state == nil {
		state = make(map[string]any)
	}
	if users == nil {
		users = make([]User, 0)
	}

	return Bundle{
		GroupRecordType: groups,
		UserRecordType:  users,
		StateRecordType: state,
	}
}

func (bundle Bundle) Groups() []Group {
	return bundle[GroupRecordType].([]Group)
}

func (bundle Bundle) Users() []User {
	return bundle[UserRecordType].([]User)
}
