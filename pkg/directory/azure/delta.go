package azure

import (
	"context"
	"errors"
	"net/url"
	"sort"

	"github.com/pomerium/datasource/pkg/directory"
)

const (
	groupsDeltaPath            = "/v1.0/groups/delta"
	servicePrincipalsDeltaPath = "/v1.0/servicePrincipals/delta"
	usersDeltaPath             = "/v1.0/users/delta"
)

type (
	deltaCollection struct {
		provider                  *Provider
		groups                    map[string]deltaGroup
		groupDeltaLink            string
		servicePrincipals         map[string]deltaServicePrincipal
		servicePrincipalDeltaLink string
		users                     map[string]deltaUser
		userDeltaLink             string
	}
	deltaGroup struct {
		ID          string                      `json:"id"`
		DisplayName string                      `json:"displayName"`
		Members     map[string]deltaGroupMember `json:"members"`
	}
	deltaGroupMember struct {
		MemberType string `json:"memberType"`
		ID         string `json:"id"`
	}
	deltaUser struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Email       string `json:"email"`
	}
	deltaServicePrincipal struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	}
)

func newDeltaCollection(p *Provider) *deltaCollection {
	return &deltaCollection{
		provider:          p,
		groups:            make(map[string]deltaGroup),
		users:             make(map[string]deltaUser),
		servicePrincipals: make(map[string]deltaServicePrincipal),
	}
}

// Sync syncs the latest changes from the microsoft graph API.
//
// Synchronization is based on https://docs.microsoft.com/en-us/graph/delta-query-groups
//
// It involves 4 steps:
//
// 1. an initial request to /v1.0/groups/delta
// 2. one or more requests to /v1.0/groups/delta?$skiptoken=..., which comes from the @odata.nextLink
// 3. a final response with @odata.deltaLink
// 4. on the next call to sync, starting at @odata.deltaLink
//
// Only the changed groups/members are returned. Removed groups/members have an @removed property.
func (dc *deltaCollection) Sync(ctx context.Context) error {
	return errors.Join(
		dc.syncGroups(ctx),
		dc.syncServicePrincipals(ctx),
		dc.syncUsers(ctx),
	)
}

func (dc *deltaCollection) syncGroups(ctx context.Context) error {
	apiURL := dc.groupDeltaLink

	// if no delta link is set yet, start the initial fill
	if apiURL == "" {
		apiURL = dc.provider.cfg.graphURL.ResolveReference(&url.URL{
			Path: groupsDeltaPath,
			RawQuery: url.Values{
				"$select": {"displayName,members"},
			}.Encode(),
		}).String()
	}

	for {
		var res groupsDeltaResponse
		err := dc.provider.api(ctx, apiURL, &res)
		if err != nil {
			return err
		}

		for _, g := range res.Value {
			// if removed exists, the group was deleted
			if g.Removed != nil {
				delete(dc.groups, g.ID)
				continue
			}

			gdg := dc.groups[g.ID]
			gdg.ID = g.ID
			gdg.DisplayName = g.DisplayName
			if gdg.Members == nil {
				gdg.Members = make(map[string]deltaGroupMember)
			}
			for _, m := range g.Members {
				// if removed exists, the member was deleted
				if m.Removed != nil {
					delete(gdg.Members, m.ID)
					continue
				}

				gdg.Members[m.ID] = deltaGroupMember{
					MemberType: m.Type,
					ID:         m.ID,
				}
			}
			dc.groups[g.ID] = gdg
		}

		switch {
		case res.NextLink != "":
			// when there's a next link we will query again
			apiURL = res.NextLink
		default:
			// once no next link is set anymore, we save the delta link and return
			dc.groupDeltaLink = res.DeltaLink
			return nil
		}
	}
}

func (dc *deltaCollection) syncServicePrincipals(ctx context.Context) error {
	apiURL := dc.servicePrincipalDeltaLink

	// if no delta link is set yet, start the initial fill
	if apiURL == "" {
		apiURL = dc.provider.cfg.graphURL.ResolveReference(&url.URL{
			Path: servicePrincipalsDeltaPath,
			RawQuery: url.Values{
				"$select": {"displayName"},
			}.Encode(),
		}).String()
	}

	for {
		var res servicePrincipalsDeltaResponse
		err := dc.provider.api(ctx, apiURL, &res)
		if err != nil {
			return err
		}

		for _, sp := range res.Value {
			// if removed exists, the service principal was deleted
			if sp.Removed != nil {
				delete(dc.servicePrincipals, sp.ID)
				continue
			}
			dc.servicePrincipals[sp.ID] = deltaServicePrincipal{
				ID:          sp.ID,
				DisplayName: sp.DisplayName,
			}
		}

		switch {
		case res.NextLink != "":
			// when there's a next link we will query again
			apiURL = res.NextLink
		default:
			// once no next link is set anymore, we save the delta link and return
			dc.servicePrincipalDeltaLink = res.DeltaLink
			return nil
		}
	}
}

func (dc *deltaCollection) syncUsers(ctx context.Context) error {
	apiURL := dc.userDeltaLink

	// if no delta link is set yet, start the initial fill
	if apiURL == "" {
		apiURL = dc.provider.cfg.graphURL.ResolveReference(&url.URL{
			Path: usersDeltaPath,
			RawQuery: url.Values{
				"$select": {"displayName,mail,userPrincipalName"},
			}.Encode(),
		}).String()
	}

	for {
		var res usersDeltaResponse
		err := dc.provider.api(ctx, apiURL, &res)
		if err != nil {
			return err
		}

		for _, u := range res.Value {
			// if removed exists, the user was deleted
			if u.Removed != nil {
				delete(dc.users, u.ID)
				continue
			}
			dc.users[u.ID] = deltaUser{
				ID:          u.ID,
				DisplayName: u.DisplayName,
				Email:       u.getEmail(),
			}
		}

		switch {
		case res.NextLink != "":
			// when there's a next link we will query again
			apiURL = res.NextLink
		default:
			// once no next link is set anymore, we save the delta link and return
			dc.userDeltaLink = res.DeltaLink
			return nil
		}
	}
}

// CurrentUserGroups returns the directory groups and users based on the current state.
func (dc *deltaCollection) CurrentUserGroups() ([]directory.Group, []directory.User) {
	var groups []directory.Group

	groupLookup := newGroupLookup()
	for _, g := range dc.groups {
		groups = append(groups, directory.Group{
			ID:   g.ID,
			Name: g.DisplayName,
		})
		var groupIDs, userIDs []string
		for _, m := range g.Members {
			switch m.MemberType {
			case "#microsoft.graph.group":
				groupIDs = append(groupIDs, m.ID)
			case "#microsoft.graph.servicePrincipal",
				"#microsoft.graph.user":
				userIDs = append(userIDs, m.ID)
			}
		}
		groupLookup.addGroup(g.ID, groupIDs, userIDs)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID < groups[j].ID
	})

	var users []directory.User
	for _, sp := range dc.servicePrincipals {
		users = append(users, directory.User{
			ID:          sp.ID,
			GroupIDs:    groupLookup.getGroupIDsForUser(sp.ID),
			DisplayName: sp.DisplayName,
		})
	}
	for _, u := range dc.users {
		users = append(users, directory.User{
			ID:          u.ID,
			GroupIDs:    groupLookup.getGroupIDsForUser(u.ID),
			DisplayName: u.DisplayName,
			Email:       u.Email,
		})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return groups, users
}

// API types for the microsoft graph API.
type (
	deltaResponseRemoved struct {
		Reason string `json:"reason"`
	}

	groupsDeltaResponse struct {
		Context   string                     `json:"@odata.context"`
		NextLink  string                     `json:"@odata.nextLink,omitempty"`
		DeltaLink string                     `json:"@odata.deltaLink,omitempty"`
		Value     []groupsDeltaResponseGroup `json:"value"`
	}
	groupsDeltaResponseGroup struct {
		apiGroup
		Members []groupsDeltaResponseGroupMember `json:"members@delta"`
		Removed *deltaResponseRemoved            `json:"@removed,omitempty"`
	}
	groupsDeltaResponseGroupMember struct {
		Type    string                `json:"@odata.type"`
		ID      string                `json:"id"`
		Removed *deltaResponseRemoved `json:"@removed,omitempty"`
	}

	servicePrincipalsDeltaResponse struct {
		Context   string                                           `json:"@odata.context"`
		NextLink  string                                           `json:"@odata.nextLink,omitempty"`
		DeltaLink string                                           `json:"@odata.deltaLink,omitempty"`
		Value     []servicePrincipalsDeltaResponseServicePrincipal `json:"value"`
	}
	servicePrincipalsDeltaResponseServicePrincipal struct {
		apiServicePrincipal
		Removed *deltaResponseRemoved `json:"@removed,omitempty"`
	}

	usersDeltaResponse struct {
		Context   string                   `json:"@odata.context"`
		NextLink  string                   `json:"@odata.nextLink,omitempty"`
		DeltaLink string                   `json:"@odata.deltaLink,omitempty"`
		Value     []usersDeltaResponseUser `json:"value"`
	}
	usersDeltaResponseUser struct {
		apiUser
		Removed *deltaResponseRemoved `json:"@removed,omitempty"`
	}
)
