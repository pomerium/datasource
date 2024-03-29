package okta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/tomnomnom/linkheader"

	"github.com/pomerium/datasource/pkg/directory"
)

// A Provider is an Okta user group directory provider.
type Provider struct {
	cfg         *config
	lastUpdated *time.Time
	groups      map[string]directory.Group
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg:    getConfig(options...),
		groups: make(map[string]directory.Group),
	}
}

// GetDirectory fetches the groups of which the user is a member
// https://developer.okta.com/docs/reference/api/users/#get-user-s-groups
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	groups, err := p.getGroups(ctx)
	if err != nil {
		return nil, nil, err
	}

	userLookup := map[string]apiUserObject{}
	userIDToGroups := map[string][]string{}
	for i := 0; i < len(groups); i++ {
		group := groups[i]
		users, err := p.getGroupMembers(ctx, group.ID)

		// if we get a 404 on the member query, it means the group doesn't exist, so we should remove it from
		// the cached lookup and the local groups list
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.HTTPStatusCode == http.StatusNotFound {
			delete(p.groups, group.ID)
			groups = append(groups[:i], groups[i+1:]...)
			i--
			continue
		}

		if err != nil {
			return nil, nil, err
		}
		for _, u := range users {
			userIDToGroups[u.ID] = append(userIDToGroups[u.ID], group.ID)
			userLookup[u.ID] = u
		}
	}

	var users []directory.User
	for _, u := range userLookup {
		groups := userIDToGroups[u.ID]
		sort.Strings(groups)
		users = append(users, directory.User{
			ID:          u.ID,
			GroupIDs:    groups,
			DisplayName: u.getDisplayName(),
			Email:       u.Profile.Email,
		})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return groups, users, nil
}

func (p *Provider) getGroups(ctx context.Context) ([]directory.Group, error) {
	groupURL, err := p.getURL("/api/v1/groups", func(q url.Values) {
		q.Set("limit", strconv.Itoa(p.cfg.batchSize))
		if p.lastUpdated != nil {
			q.Set("filter", fmt.Sprintf(`lastUpdated gt "%[1]s" or lastMembershipUpdated gt "%[1]s"`, p.lastUpdated.UTC().Format(filterDateFormat)))
		} else {
			now := time.Now()
			p.lastUpdated = &now
		}
	})
	if err != nil {
		return nil, err
	}

	for groupURL != "" {
		var out []apiGroupObject
		hdrs, err := p.apiGet(ctx, groupURL, &out)
		if err != nil {
			return nil, fmt.Errorf("okta: error querying for groups: %w", err)
		}

		for _, el := range out {
			lu, _ := time.Parse(el.LastUpdated, filterDateFormat)
			lmu, _ := time.Parse(el.LastMembershipUpdated, filterDateFormat)
			if lu.After(*p.lastUpdated) {
				p.lastUpdated = &lu
			}
			if lmu.After(*p.lastUpdated) {
				p.lastUpdated = &lmu
			}
			p.groups[el.ID] = directory.Group{
				ID:   el.ID,
				Name: el.Profile.Name,
			}
		}
		groupURL = getNextLink(hdrs)
	}

	groups := make([]directory.Group, 0, len(p.groups))
	for _, dg := range p.groups {
		groups = append(groups, dg)
	}
	return groups, nil
}

func (p *Provider) getGroupMembers(ctx context.Context, groupID string) (users []apiUserObject, err error) {
	usersURL, err := p.getURL(fmt.Sprintf("/api/v1/groups/%s/users", groupID), func(qs url.Values) {
		qs.Set("limit", fmt.Sprint(p.cfg.batchSize))
	})
	if err != nil {
		return nil, err
	}

	for usersURL != "" {
		var out []apiUserObject
		hdrs, err := p.apiGet(ctx, usersURL, &out)
		if err != nil {
			return nil, fmt.Errorf("okta: error querying for groups: %w", err)
		}

		users = append(users, out...)
		usersURL = getNextLink(hdrs)
	}

	return users, nil
}

func (p *Provider) apiGet(ctx context.Context, uri string, out interface{}) (http.Header, error) {
	apiKey := p.cfg.apiKey
	if apiKey == "" {
		return nil, ErrAPIKeyRequired
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("okta: failed to create HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "SSWS "+apiKey)

	for {
		res, err := p.cfg.getHTTPClient().Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusTooManyRequests {
			limitReset, err := strconv.ParseInt(res.Header.Get("X-Rate-Limit-Reset"), 10, 64)
			if err == nil {
				time.Sleep(time.Until(time.Unix(limitReset, 0)))
			}
			continue
		}
		if res.StatusCode/100 != httpSuccessClass {
			return nil, newAPIError(res)
		}
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return nil, err
		}
		return res.Header, nil
	}
}

func (p *Provider) getURL(path string, modifyQuery func(url.Values)) (string, error) {
	if p.cfg.url == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(p.cfg.url)
	if err != nil {
		return "", ErrInvalidURL
	}
	q := u.Query()
	modifyQuery(q)
	u = u.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: q.Encode(),
	})
	return u.String(), nil
}

func getNextLink(hdrs http.Header) string {
	for _, link := range linkheader.ParseMultiple(hdrs.Values("Link")) {
		if link.Rel == "next" {
			return link.URL
		}
	}
	return ""
}

// An APIError is an error from the okta API.
type APIError struct {
	HTTPStatusCode int
	Body           string
	ErrorCode      string   `json:"errorCode"`
	ErrorSummary   string   `json:"errorSummary"`
	ErrorLink      string   `json:"errorLink"`
	ErrorID        string   `json:"errorId"`
	ErrorCauses    []string `json:"errorCauses"`
}

func newAPIError(res *http.Response) error {
	if res == nil {
		return nil
	}
	buf, _ := io.ReadAll(io.LimitReader(res.Body, readLimit)) // limit to 100kb

	err := &APIError{
		HTTPStatusCode: res.StatusCode,
		Body:           string(buf),
	}
	_ = json.Unmarshal(buf, err)
	return err
}

func (err *APIError) Error() string {
	return fmt.Sprintf("okta: error querying API, status_code=%d: %s", err.HTTPStatusCode, err.Body)
}

type (
	apiGroupObject struct {
		ID      string `json:"id"`
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
		LastUpdated           string `json:"lastUpdated"`
		LastMembershipUpdated string `json:"lastMembershipUpdated"`
	}
	apiUserObject struct {
		ID      string `json:"id"`
		Profile struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Email     string `json:"email"`
		} `json:"profile"`
	}
)

func (obj *apiUserObject) getDisplayName() string {
	return obj.Profile.FirstName + " " + obj.Profile.LastName
}
