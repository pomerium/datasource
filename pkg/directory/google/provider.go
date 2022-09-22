package google

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/pomerium/datasource/pkg/directory"
)

const currentAccountCustomerID = "my_customer"

// Required scopes for groups api
// https://developers.google.com/admin-sdk/directory/v1/reference/groups/list
var apiScopes = []string{admin.AdminDirectoryUserReadonlyScope, admin.AdminDirectoryGroupReadonlyScope}

// A Provider is a Google directory provider.
type Provider struct {
	cfg *config

	mu        sync.RWMutex
	apiClient *admin.Service
}

// New creates a new Google directory provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),
	}
}

// GetDirectory returns a slice of group names a given user is in
// NOTE: groups via Directory API is limited to 1 QPS!
// https://developers.google.com/admin-sdk/directory/v1/reference/groups/list
// https://developers.google.com/admin-sdk/directory/v1/limits
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	apiClient, err := p.getAPIClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("google: error getting API client: %w", err)
	}

	// query all the groups
	var groups []directory.Group
	err = apiClient.Groups.List().
		Context(ctx).
		Customer(currentAccountCustomerID).
		Pages(ctx, func(res *admin.Groups) error {
			for _, g := range res.Groups {
				// Skip group without member.
				if g.DirectMembersCount == 0 {
					continue
				}
				groups = append(groups, directory.Group{
					ID:    g.Id,
					Name:  g.Email,
					Email: g.Email,
				})
			}
			return nil
		})
	if err != nil {
		return nil, nil, fmt.Errorf("google: error getting groups: %w", err)
	}

	// query all the user members for each group
	// - create a lookup table for the user (storing id and name)
	//   (this includes users who aren't necessarily members of the same organization)
	// - create a lookup table for the user's groups
	userLookup := map[string]apiUserObject{}
	userIDToGroups := map[string][]string{}
	for _, group := range groups {
		group := group
		err = apiClient.Members.List(group.ID).
			Context(ctx).
			Pages(ctx, func(res *admin.Members) error {
				for _, member := range res.Members {
					// only include user objects
					if member.Type != "USER" {
						continue
					}

					userLookup[member.Id] = apiUserObject{
						ID:    member.Id,
						Email: member.Email,
					}
					userIDToGroups[member.Id] = append(userIDToGroups[member.Id], group.ID)
				}
				return nil
			})
		if err != nil {
			return nil, nil, fmt.Errorf("google: error getting group members: %w", err)
		}
	}

	// query all the users in the organization
	err = apiClient.Users.List().
		Context(ctx).
		Customer(currentAccountCustomerID).
		Pages(ctx, func(res *admin.Users) error {
			for _, u := range res.Users {
				auo := apiUserObject{
					ID:    u.Id,
					Email: u.PrimaryEmail,
				}
				if u.Name != nil {
					auo.DisplayName = u.Name.FullName
				}
				userLookup[u.Id] = auo
			}
			return nil
		})
	if err != nil {
		return nil, nil, fmt.Errorf("google: error getting users: %w", err)
	}

	var users []directory.User
	for _, u := range userLookup {
		groups := userIDToGroups[u.ID]
		sort.Strings(groups)
		users = append(users, directory.User{
			ID:          u.ID,
			GroupIDs:    groups,
			DisplayName: u.DisplayName,
			Email:       u.Email,
		})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return groups, users, nil
}

func (p *Provider) getAPIClient(ctx context.Context) (*admin.Service, error) {
	jsonKey := p.cfg.jsonKey
	if len(jsonKey) == 0 {
		return nil, ErrJSONKeyRequired
	}
	impersonateUser := p.cfg.impersonateUser
	if impersonateUser == "" {
		return nil, ErrImpersonateUserRequired
	}

	p.mu.RLock()
	apiClient := p.apiClient
	p.mu.RUnlock()
	if apiClient != nil {
		return apiClient, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.apiClient != nil {
		return p.apiClient, nil
	}

	config, err := google.JWTConfigFromJSON(jsonKey, apiScopes...)
	if err != nil {
		return nil, fmt.Errorf("google: error reading jwt config: %w", err)
	}
	config.Subject = impersonateUser

	ts := config.TokenSource(ctx)

	p.apiClient, err = admin.NewService(ctx,
		option.WithHTTPClient(p.cfg.httpClient),
		option.WithTokenSource(ts),
		option.WithEndpoint(p.cfg.url))
	if err != nil {
		return nil, fmt.Errorf("google: failed creating admin service %w", err)
	}
	return p.apiClient, nil
}

type apiUserObject struct {
	ID          string
	DisplayName string
	Email       string
}
