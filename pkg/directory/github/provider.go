package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

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

	userLoginToGroups := map[string][]string{}

	var allGroups []directory.Group
	for _, orgSlug := range orgSlugs {
		teams, err := p.listOrganizationTeamsWithMemberIDs(ctx, orgSlug)
		if err != nil {
			return nil, nil, err
		}

		for _, team := range teams {
			allGroups = append(allGroups, directory.Group{
				ID:   team.Slug,
				Name: team.Slug,
			})
			for _, memberID := range team.MemberIDs {
				userLoginToGroups[memberID] = append(userLoginToGroups[memberID], team.Slug)
			}
		}
	}
	sort.Slice(allGroups, func(i, j int) bool {
		return allGroups[i].ID < allGroups[j].ID
	})

	var allUsers []directory.User
	for _, orgSlug := range orgSlugs {
		members, err := p.listOrganizationMembers(ctx, orgSlug)
		if err != nil {
			return nil, nil, err
		}

		for _, member := range members {
			du := directory.User{
				ID:          member.Login,
				GroupIDs:    userLoginToGroups[member.ID],
				DisplayName: member.Name,
				Email:       member.Email,
			}
			sort.Strings(du.GroupIDs)
			allUsers = append(allUsers, du)
		}
	}
	sort.Slice(allUsers, func(i, j int) bool {
		return allUsers[i].ID < allUsers[j].ID
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

func (p *Provider) api(ctx context.Context, apiURL string, out interface{}) (http.Header, error) {
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

	res, err := p.cfg.httpClient.Do(req)
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

func (p *Provider) graphql(ctx context.Context, query string, out interface{}) (http.Header, error) {
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

	res, err := p.cfg.httpClient.Do(req)
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
