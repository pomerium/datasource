package okta

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
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
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	ctx, client, err := okta.NewClient(ctx,
		okta.WithHttpClientPtr(p.cfg.getHTTPClient()),
		okta.WithOrgUrl(p.cfg.url),
		okta.WithTestingDisableHttpsCheck(true),
		okta.WithToken(p.cfg.apiKey))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating okta client: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.sync(ctx, client)
	if err != nil {
		return nil, nil, err
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
	return groups, users, nil
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
		p.groups = make(map[string]okta.Group)
		p.groupMembers = make(map[string][]okta.User)

		var err error
		changedGroups, _, err = client.Group.ListGroups(ctx, &query.Params{
			Limit: int64(p.cfg.batchSize),
		})
		if err != nil {
			return fmt.Errorf("error querying all groups: %w", err)
		}
	} else {
		// sync changes
		var err error
		filter := fmt.Sprintf(`lastUpdated gt "%s" or lastMembershipUpdated gt "%s"`,
			lastUpdated.UTC().Format(filterDateFormat),
			lastMembershipUpdated.UTC().Format(filterDateFormat))
		changedGroups, _, err = client.Group.ListGroups(ctx, &query.Params{
			Limit:  int64(p.cfg.batchSize),
			Filter: filter,
		})
		if err != nil {
			return fmt.Errorf("error querying changed groups: %w", err)
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

		users, _, err := client.Group.ListGroupUsers(ctx, g.Id, &query.Params{
			Limit: int64(p.cfg.batchSize),
		})
		if err != nil {
			return fmt.Errorf("error listing group members: %w", err)
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
