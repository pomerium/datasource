package okta

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"golang.org/x/exp/maps"

	"github.com/pomerium/datasource/pkg/directory"
)

// A Provider is an Okta user group directory provider.
type Provider struct {
	cfg *config

	mu           sync.Mutex
	groups       map[string]okta.Group
	groupMembers map[string][]okta.User
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),

		groups:       make(map[string]okta.Group),
		groupMembers: make(map[string][]okta.User),
	}
}

// GetDirectory gets the full directory information for Okta.
func (p *Provider) GetDirectory(ctx context.Context) (directory.Bundle, error) {
	ctx, client, err := okta.NewClient(ctx,
		append([]okta.ConfigSetter{
			okta.WithHttpClientPtr(p.cfg.getHTTPClient()),
			okta.WithOrgUrl(p.cfg.url),
			okta.WithToken(p.cfg.apiKey),
		}, p.cfg.oktaOptions...)...)
	if err != nil {
		return directory.Bundle{}, fmt.Errorf("error creating okta client: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.sync(ctx, client)
	if err != nil {
		return directory.Bundle{}, err
	}

	var groups []directory.Group
	userLookup := map[string]directory.User{}
	for _, g := range p.groups {
		groups = append(groups, directory.Group{
			ID:   g.Id,
			Name: g.Profile.Name,
		})
		for _, u := range p.groupMembers[g.Id] {
			du := userLookup[u.Id]
			du.DisplayName = getUserDisplayName(u)
			du.Email = getUserEmail(u)
			du.GroupIDs = append(du.GroupIDs, g.Id)
			sort.Strings(du.GroupIDs)
			du.ID = u.Id
			userLookup[u.Id] = du
		}
	}

	users := maps.Values(userLookup)

	// sort groups and users
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID < groups[j].ID
	})
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return directory.NewBundle(groups, users, nil), nil
}

func (p *Provider) sync(ctx context.Context, client *okta.Client) error {
	var lastMembershipUpdated, lastUpdated time.Time
	for _, g := range p.groups {
		if g.LastMembershipUpdated != nil && g.LastMembershipUpdated.After(lastMembershipUpdated) {
			lastMembershipUpdated = *g.LastMembershipUpdated
		}
		if g.LastUpdated != nil && g.LastUpdated.After(lastUpdated) {
			lastUpdated = *g.LastUpdated
		}
	}

	var changedGroups []*okta.Group
	if lastMembershipUpdated.IsZero() || lastUpdated.IsZero() {
		// full sync
		clear(p.groups)
		clear(p.groupMembers)

		var err error
		changedGroups, err = listAllGroups(ctx, client, p.cfg.batchSize)
		if err != nil {
			return err
		}
	} else {
		// sync changes
		var err error
		changedGroups, err = listChangedGroups(ctx, client, lastUpdated, lastMembershipUpdated, p.cfg.batchSize)
		if err != nil {
			return err
		}
	}

	err := p.syncGroups(ctx, client, changedGroups)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) syncGroups(ctx context.Context, client *okta.Client, groups []*okta.Group) error {
	for _, g := range groups {
		p.groups[g.Id] = *g

		users, err := listGroupUsers(ctx, client, g.Id, p.cfg.batchSize)
		if err != nil {
			return err
		}
		s := make([]okta.User, len(users))
		for i, u := range users {
			s[i] = *u
		}
		p.groupMembers[g.Id] = s
	}

	return nil
}

func getUserDisplayName(user okta.User) string {
	if user.Profile == nil {
		return ""
	}

	firstName, _ := (*user.Profile)["firstName"].(string)
	lastName, _ := (*user.Profile)["lastName"].(string)
	return firstName + " " + lastName
}

func getUserEmail(user okta.User) string {
	if user.Profile == nil {
		return ""
	}

	email, ok := (*user.Profile)["email"].(string)
	if !ok {
		return ""
	}

	return email
}
