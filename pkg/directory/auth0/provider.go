package auth0

import (
	"context"
	"fmt"
	"sort"

	"gopkg.in/auth0.v5/management"

	"github.com/pomerium/datasource/pkg/directory"
)

// Provider is an Auth0 user group directory provider.
type Provider struct {
	cfg *config
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),
	}
}

func (p *Provider) getManagers(ctx context.Context) (RoleManager, UserManager, error) {
	return p.cfg.newManagers(ctx, p.cfg)
}

// UserGroups fetches a slice of groups and users.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	rm, _, err := p.getManagers(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: could not get the role manager: %w", err)
	}

	roles, err := getRoles(rm)
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: %w", err)
	}

	userIDToGroups := map[string][]string{}
	for _, role := range roles {
		ids, err := getRoleUserIDs(rm, role.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("auth0: %w", err)
		}

		for _, id := range ids {
			userIDToGroups[id] = append(userIDToGroups[id], role.ID)
		}
	}

	var users []directory.User
	for userID, groups := range userIDToGroups {
		sort.Strings(groups)
		users = append(users, directory.User{
			ID:       userID,
			GroupIDs: groups,
		})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return roles, users, nil
}

func getRoles(rm RoleManager) ([]directory.Group, error) {
	roles := []directory.Group{}

	shouldContinue := true
	page := 0

	for shouldContinue {
		listRes, err := rm.List(management.IncludeTotals(true), management.Page(page))
		if err != nil {
			return nil, fmt.Errorf("could not list roles: %w", err)
		}

		for _, role := range listRes.Roles {
			roles = append(roles, directory.Group{
				ID:   *role.ID,
				Name: *role.Name,
			})
		}

		page++
		shouldContinue = listRes.HasNext()
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID < roles[j].ID
	})
	return roles, nil
}

func getRoleUserIDs(rm RoleManager, roleID string) ([]string, error) {
	var ids []string

	shouldContinue := true
	page := 0

	for shouldContinue {
		usersRes, err := rm.Users(roleID, management.IncludeTotals(true), management.Page(page))
		if err != nil {
			return nil, fmt.Errorf("could not get users for role %q: %w", roleID, err)
		}

		for _, user := range usersRes.Users {
			ids = append(ids, *user.ID)
		}

		page++
		shouldContinue = usersRes.HasNext()
	}

	sort.Strings(ids)
	return ids, nil
}
