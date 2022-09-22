package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/tomnomnom/linkheader"

	"github.com/pomerium/datasource/pkg/directory"
)

// The Provider retrieves users and groups from gitlab.
type Provider struct {
	cfg *config
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),
	}
}

// UserGroups gets the directory user groups for gitlab.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	groups, err := p.listGroups(ctx)
	if err != nil {
		return nil, nil, err
	}

	userLookup := map[int]apiUserObject{}
	userIDToGroupIDs := map[int][]string{}
	for _, group := range groups {
		users, err := p.listGroupMembers(ctx, group.ID)
		if err != nil {
			return nil, nil, err
		}

		for _, u := range users {
			userIDToGroupIDs[u.ID] = append(userIDToGroupIDs[u.ID], group.ID)
			userLookup[u.ID] = u
		}
	}

	var users []directory.User
	for _, u := range userLookup {
		user := directory.User{
			ID:          fmt.Sprint(u.ID),
			DisplayName: u.Name,
			Email:       u.Email,
		}

		user.GroupIDs = append(user.GroupIDs, userIDToGroupIDs[u.ID]...)

		sort.Strings(user.GroupIDs)
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return groups, users, nil
}

// listGroups returns a map, with key is group ID, element is group name.
func (p *Provider) listGroups(ctx context.Context) ([]directory.Group, error) {
	nextURL := p.cfg.url.ResolveReference(&url.URL{
		Path: "/api/v4/groups",
	}).String()
	var groups []directory.Group
	for nextURL != "" {
		var result []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		hdrs, err := p.api(ctx, nextURL, &result)
		if err != nil {
			return nil, fmt.Errorf("gitlab: error querying groups: %w", err)
		}

		for _, r := range result {
			groups = append(groups, directory.Group{
				ID:   strconv.Itoa(r.ID),
				Name: r.Name,
			})
		}

		nextURL = getNextLink(hdrs)
	}
	return groups, nil
}

func (p *Provider) listGroupMembers(ctx context.Context, groupID string) (users []apiUserObject, err error) {
	nextURL := p.cfg.url.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/api/v4/groups/%s/members", groupID),
	}).String()
	for nextURL != "" {
		var result []apiUserObject
		hdrs, err := p.api(ctx, nextURL, &result)
		if err != nil {
			return nil, fmt.Errorf("gitlab: error querying group members: %w", err)
		}

		users = append(users, result...)
		nextURL = getNextLink(hdrs)
	}
	return users, nil
}

func (p *Provider) api(ctx context.Context, uri string, out interface{}) (http.Header, error) {
	privateToken := p.cfg.privateToken
	if privateToken == "" {
		return nil, ErrPrivateTokenRequired
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("gitlab: failed to create HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", privateToken)

	res, err := p.cfg.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("gitlab: error querying api url=%s status_code=%d: %s", uri, res.StatusCode, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		return nil, err
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

type apiUserObject struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
