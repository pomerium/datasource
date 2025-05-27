package okta

import (
	"context"
	"encoding/json"

	"github.com/okta/okta-sdk-golang/v2/okta"
)

type directoryState struct {
	Groups       map[string]okta.Group  `json:"groups,omitempty"`
	GroupMembers map[string][]okta.User `json:"groupMembers,omitempty"`
}

func (p *Provider) GetDirectoryState(_ context.Context) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return json.Marshal(directoryState{
		Groups:       p.groups,
		GroupMembers: p.groupMembers,
	})
}

func (p *Provider) SetDirectoryState(_ context.Context, state []byte) error {
	var ds directoryState
	err := json.Unmarshal(state, &ds)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.groups = ds.Groups
	p.groupMembers = ds.GroupMembers
	p.mu.Unlock()

	return nil
}
