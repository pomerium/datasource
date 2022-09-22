package onelogin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/oauth2"

	"github.com/pomerium/datasource/pkg/directory"
)

// The Provider retrieves users and groups from onelogin.
type Provider struct {
	cfg *config

	mu    sync.RWMutex
	token *oauth2.Token
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	cfg := getConfig(options...)
	return &Provider{
		cfg: cfg,
	}
}

// GetDirectory gets the directory user groups for onelogin.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, nil, err
	}

	groups, err := p.listGroups(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	apiUsers, err := p.listUsers(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	var users []directory.User
	for _, u := range apiUsers {
		users = append(users, directory.User{
			ID:          strconv.Itoa(u.ID),
			GroupIDs:    []string{strconv.Itoa(u.GroupID)},
			DisplayName: u.FirstName + " " + u.LastName,
			Email:       u.Email,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return groups, users, nil
}

func (p *Provider) listGroups(ctx context.Context, accessToken string) ([]directory.Group, error) {
	var groups []directory.Group
	apiURL := p.cfg.apiURL.ResolveReference(&url.URL{
		Path:     "/api/1/groups",
		RawQuery: fmt.Sprintf("limit=%d", p.cfg.batchSize),
	}).String()
	for apiURL != "" {
		var result []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		nextLink, err := p.apiGet(ctx, accessToken, apiURL, &result)
		if err != nil {
			return nil, fmt.Errorf("onelogin: listing groups: %w", err)
		}

		for _, r := range result {
			groups = append(groups, directory.Group{
				ID:   strconv.Itoa(r.ID),
				Name: r.Name,
			})
		}

		apiURL = nextLink
	}
	return groups, nil
}

func (p *Provider) listUsers(ctx context.Context, accessToken string) ([]apiUserObject, error) {
	var users []apiUserObject

	apiURL := p.cfg.apiURL.ResolveReference(&url.URL{
		Path:     "/api/1/users",
		RawQuery: fmt.Sprintf("limit=%d", p.cfg.batchSize),
	}).String()
	for apiURL != "" {
		var result []apiUserObject
		nextLink, err := p.apiGet(ctx, accessToken, apiURL, &result)
		if err != nil {
			return nil, fmt.Errorf("onelogin: listing users: %w", err)
		}

		users = append(users, result...)
		apiURL = nextLink
	}

	return users, nil
}

func (p *Provider) apiGet(ctx context.Context, accessToken string, uri string, out interface{}) (nextLink string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("bearer:%s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := p.cfg.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return "", fmt.Errorf("onelogin: error querying api: %s", res.Status)
	}

	var result struct {
		Pagination struct {
			NextLink string `json:"next_link"`
		}
		Data json.RawMessage `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(result.Data, out)
	if err != nil {
		return "", err
	}

	return result.Pagination.NextLink, nil
}

func (p *Provider) getToken(ctx context.Context) (*oauth2.Token, error) {
	clientID := p.cfg.clientID
	if clientID == "" {
		return nil, ErrClientIDRequired
	}
	clientSecret := p.cfg.clientSecret
	if clientSecret == "" {
		return nil, ErrClientSecretRequired
	}

	p.mu.RLock()
	token := p.token
	p.mu.RUnlock()

	if token != nil && token.Valid() {
		return token, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	token = p.token
	if token != nil && token.Valid() {
		return token, nil
	}

	apiURL := p.cfg.apiURL.ResolveReference(&url.URL{
		Path: "/auth/oauth2/v2/token",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL.String(), strings.NewReader(`{ "grant_type": "client_credentials" }`))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("client_id:%s, client_secret:%s",
		clientID, clientSecret))
	req.Header.Set("Content-Type", "application/json")

	res, err := p.cfg.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("onelogin: error querying oauth2 token: %s", res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return nil, err
	}
	p.token = token

	return p.token, nil
}

type apiUserObject struct {
	ID        int    `json:"id"`
	GroupID   int    `json:"group_id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}
