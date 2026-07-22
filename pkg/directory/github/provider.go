package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/tomnomnom/linkheader"

	"github.com/pomerium/datasource/pkg/directory"
)

// The Provider retrieves users and groups from github.
type Provider struct {
	cfg *config
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),
	}
}

// GetDirectory gets the directory user groups for github.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	orgSlugs, err := p.listOrgs(ctx)
	if err != nil {
		return nil, nil, err
	}

	userNodeIDToGroups := map[string][]string{}

	groupLookup := map[string]directory.Group{}
	for _, orgSlug := range orgSlugs {
		teams, err := p.listOrganizationTeamsWithMemberIDs(ctx, orgSlug)
		if err != nil {
			return nil, nil, err
		}

		for _, team := range teams {
			dg := directory.Group{
				Name: team.Slug,
			}
			if p.cfg.useNodeIDs {
				dg.ID = team.NodeID
			} else {
				dg.ID = team.Slug
			}
			groupLookup[dg.ID] = dg
			for _, memberNodeID := range team.MemberNodeIDs {
				userNodeIDToGroups[memberNodeID] = append(userNodeIDToGroups[memberNodeID], dg.ID)
			}
		}
	}

	userLookup := map[string]directory.User{}
	for _, orgSlug := range orgSlugs {
		members, err := p.listOrganizationMembers(ctx, orgSlug)
		if err != nil {
			return nil, nil, err
		}

		for _, member := range members {
			du := directory.User{
				GroupIDs:    userNodeIDToGroups[member.NodeID],
				DisplayName: member.Name,
				Email:       member.Email,
			}
			if p.cfg.useNodeIDs {
				du.ID = member.NodeID
			} else {
				du.ID = member.Login
			}
			sort.Strings(du.GroupIDs)
			userLookup[du.ID] = du
		}
	}

	allGroups := slices.SortedFunc(maps.Values(groupLookup), func(dg1, dg2 directory.Group) int {
		return strings.Compare(dg1.ID, dg2.ID)
	})
	allUsers := slices.SortedFunc(maps.Values(userLookup), func(du1, du2 directory.User) int {
		return strings.Compare(du1.ID, du2.ID)
	})

	return allGroups, allUsers, nil
}

func (p *Provider) listOrgs(ctx context.Context) (orgSlugs []string, err error) {
	nextURL := p.cfg.url.ResolveReference(&url.URL{
		Path: "/user/orgs",
	}).String()

	for nextURL != "" {
		var results []struct {
			Login string `json:"login"`
		}
		hdrs, err := p.api(ctx, nextURL, &results)
		if err != nil {
			return nil, err
		}

		for _, result := range results {
			orgSlugs = append(orgSlugs, result.Login)
		}

		nextURL = getNextLink(hdrs)
	}

	return orgSlugs, nil
}

func (p *Provider) api(ctx context.Context, apiURL string, out any) (http.Header, error) {
	username := p.cfg.username
	if username == "" {
		return nil, ErrUsernameRequired
	}
	personalAccessToken := p.cfg.personalAccessToken
	if personalAccessToken == "" {
		return nil, ErrPersonalAccessTokenRequired
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create http request: %w", err)
	}
	req.SetBasicAuth(username, personalAccessToken)

	res, err := p.cfg.getHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: failed to make http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("github: error from API: %s", res.Status)
	}

	if out != nil {
		err := json.NewDecoder(res.Body).Decode(out)
		if err != nil {
			return nil, fmt.Errorf("github: failed to decode json body: %w", err)
		}
	}

	return res.Header, nil
}

func (p *Provider) graphql(ctx context.Context, query string, out any) (http.Header, error) {
	username := p.cfg.username
	if username == "" {
		return nil, ErrUsernameRequired
	}
	personalAccessToken := p.cfg.personalAccessToken
	if personalAccessToken == "" {
		return nil, ErrPersonalAccessTokenRequired
	}

	apiURL := p.cfg.url.ResolveReference(&url.URL{
		Path: "/graphql",
	}).String()

	bs, _ := json.Marshal(struct {
		Query string `json:"query"`
	}{query})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("github: failed to create http request: %w", err)
	}
	req.SetBasicAuth(username, personalAccessToken)

	res, err := p.cfg.getHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: failed to make http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("github: error from API: %s", res.Status)
	}

	if out != nil {
		err := json.NewDecoder(res.Body).Decode(out)
		if err != nil {
			return nil, fmt.Errorf("github: failed to decode json body: %w", err)
		}
	}

	return res.Header, nil
}

func getNextLink(hdrs http.Header) string {
	for _, link := range linkheader.ParseMultiple(hdrs.Values("Link")) {
		if link.Rel == "next" {
			return link.URL
		}
	}
	return ""
}
