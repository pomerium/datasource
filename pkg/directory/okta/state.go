package okta

import (
	"context"
	"encoding/json"
	"io"

	"github.com/okta/okta-sdk-golang/v2/okta"

	"github.com/pomerium/datasource/pkg/directory"
)

type directoryState struct {
	Groups       map[string]okta.Group  `json:"groups,omitempty"`
	GroupMembers map[string][]okta.User `json:"groupMembers,omitempty"`
}

var _ directory.PersistentProvider = (*Provider)(nil)

// SaveDirectoryState saves the directory state to a writer.
func (p *Provider) SaveDirectoryState(_ context.Context, dst io.Writer) error {
	return json.NewEncoder(dst).Encode(directoryState{
		Groups:       p.groups,
		GroupMembers: p.groupMembers,
	})
}

// LoadDirectoryState loads the directory state from a reader.
func (p *Provider) LoadDirectoryState(_ context.Context, src io.Reader) error {
	var ds directoryState
	err := json.NewDecoder(src).Decode(&ds)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.groups = ds.Groups
	p.groupMembers = ds.GroupMembers
	p.mu.Unlock()

	return nil
}
