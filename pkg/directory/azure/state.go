package azure

import (
	"context"
	"encoding/json"
	"io"

	"github.com/pomerium/datasource/pkg/directory"
)

var _ directory.PersistentProvider = (*Provider)(nil)

type directoryState struct {
	GroupDeltaLink            string                           `json:"groupDeltaLink,omitempty"`
	Groups                    map[string]deltaGroup            `json:"groups,omitempty"`
	ServicePrincipalDeltaLink string                           `json:"servicePrincipalDeltaLink,omitempty"`
	ServicePrincipals         map[string]deltaServicePrincipal `json:"servicePrincipals,omitempty"`
	UserDeltaLink             string                           `json:"userDeltaLink,omitempty"`
	Users                     map[string]deltaUser             `json:"users,omitempty"`
}

// SaveDirectoryState saves the directory state to a writer.
func (p *Provider) SaveDirectoryState(_ context.Context, dst io.Writer) error {
	return json.NewEncoder(dst).Encode(directoryState{
		GroupDeltaLink:            p.dc.groupDeltaLink,
		Groups:                    p.dc.groups,
		ServicePrincipalDeltaLink: p.dc.servicePrincipalDeltaLink,
		ServicePrincipals:         p.dc.servicePrincipals,
		UserDeltaLink:             p.dc.userDeltaLink,
		Users:                     p.dc.users,
	})
}

// LoadDirectoryState loads the directory state from a reader.
func (p *Provider) LoadDirectoryState(_ context.Context, src io.Reader) error {
	var ds directoryState
	err := json.NewDecoder(src).Decode(&ds)
	if err != nil {
		return err
	}

	p.dc.groupDeltaLink = ds.GroupDeltaLink
	p.dc.groups = ds.Groups
	if p.dc.groups == nil {
		p.dc.groups = make(map[string]deltaGroup)
	}

	p.dc.servicePrincipalDeltaLink = ds.ServicePrincipalDeltaLink
	p.dc.servicePrincipals = ds.ServicePrincipals
	if p.dc.servicePrincipals == nil {
		p.dc.servicePrincipals = make(map[string]deltaServicePrincipal)
	}

	p.dc.userDeltaLink = ds.UserDeltaLink
	p.dc.users = ds.Users
	if p.dc.users == nil {
		p.dc.users = make(map[string]deltaUser)
	}

	return nil
}
