package auth0

import (
	"cmp"
	"context"
	"fmt"
	"iter"
	"maps"
	"slices"

	"github.com/auth0/go-auth0/management"

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

// UserGroups fetches a slice of groups and users.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	m, err := p.getManagement()
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: error creating management client: %w", err)
	}

	users := map[string]directory.User{}
	for u, err := range listUsers(ctx, m) {
		if err != nil {
			return nil, nil, fmt.Errorf("auth0: error listing users: %w", err)
		}

		users[u.GetID()] = directory.User{
			ID:          u.GetID(),
			DisplayName: u.GetName(),
			Email:       u.GetEmail(),
		}
	}

	var groups []directory.Group
	for role, err := range listRoles(ctx, m) {
		if err != nil {
			return nil, nil, fmt.Errorf("auth0: error listing roles: %w", err)
		}

		groups = append(groups, directory.Group{
			ID:   role.GetID(),
			Name: role.GetName(),
		})

		for id, err := range listRoleUserIDs(ctx, m, role.GetID()) {
			if err != nil {
				return nil, nil, fmt.Errorf("auth0: error listing role users: %w", err)
			}
			u, ok := users[id]
			if !ok {
				continue
			}
			u.GroupIDs = append(u.GroupIDs, role.GetID())
			users[id] = u
		}
	}

	// sort for determinism
	for _, u := range users {
		slices.Sort(u.GroupIDs)
	}
	return slices.SortedFunc(slices.Values(groups), func(g1, g2 directory.Group) int {
			return cmp.Compare(g1.ID, g2.ID)
		}), slices.SortedFunc(maps.Values(users), func(u1, u2 directory.User) int {
			return cmp.Compare(u1.ID, u2.ID)
		}), nil
}

func (p *Provider) getManagement() (*management.Management, error) {
	options := []management.Option{
		management.WithClientCredentials(context.Background(), p.cfg.clientID, p.cfg.clientSecret),
		management.WithClient(p.cfg.getHTTPClient()),
	}
	if p.cfg.insecure {
		options = append(options, management.WithInsecure())
	}
	return management.New(p.cfg.domain, options...)
}

func listRoleUserIDs(ctx context.Context, m *management.Management, roleID string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for page := 0; ; page++ {
			res, err := m.Role.Users(ctx, roleID, management.Page(page), management.PerPage(100))
			if err != nil {
				if !yield("", err) {
					return
				}
			}

			for _, v := range res.Users {
				if !yield(v.GetID(), nil) {
					return
				}
			}

			if !res.HasNext() {
				break
			}
		}
	}
}

func listRoles(ctx context.Context, m *management.Management) iter.Seq2[*management.Role, error] {
	return func(yield func(*management.Role, error) bool) {
		for page := 0; ; page++ {
			res, err := m.Role.List(ctx, management.Page(page), management.PerPage(100))
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			for _, v := range res.Roles {
				if !yield(v, nil) {
					return
				}
			}

			if !res.HasNext() {
				break
			}
		}
	}
}

func listUsers(ctx context.Context, m *management.Management) iter.Seq2[*management.User, error] {
	return func(yield func(*management.User, error) bool) {
		for page := 0; ; page++ {
			res, err := m.User.List(ctx, management.Page(page), management.PerPage(100))
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			for _, v := range res.Users {
				if !yield(v, nil) {
					return
				}
			}

			if !res.HasNext() {
				break
			}
		}
	}
}
