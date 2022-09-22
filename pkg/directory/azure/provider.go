package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/oauth2"

	"github.com/pomerium/datasource/pkg/directory"
)

// A Provider is a directory implementation using azure active directory.
type Provider struct {
	cfg *config
	dc  *deltaCollection

	mu    sync.RWMutex
	token *oauth2.Token
}

// New creates a new Provider.
func New(options ...Option) *Provider {
	p := &Provider{
		cfg: getConfig(options...),
	}
	p.dc = newDeltaCollection(p)
	return p
}

// GetDirectory returns the directory users in azure active directory.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	err := p.dc.Sync(ctx)
	if err != nil {
		return nil, nil, err
	}

	groups, users := p.dc.CurrentUserGroups()
	return groups, users, nil
}

func (p *Provider) api(ctx context.Context, url string, out interface{}) error {
	token, err := p.getToken(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("azure: error creating HTTP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := p.cfg.getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("azure: error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	// if we get unauthorized, invalidate the token
	if res.StatusCode == http.StatusUnauthorized {
		p.mu.Lock()
		p.token = nil
		p.mu.Unlock()
	}

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("azure: error querying api (%s): %s", url, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		return fmt.Errorf("azure: error decoding api response: %w", err)
	}

	return nil
}

func (p *Provider) getToken(ctx context.Context) (*oauth2.Token, error) {
	directoryID := p.cfg.directoryID
	if directoryID == "" {
		return nil, ErrDirectoryIDRequired
	}
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

	tokenURL := p.cfg.loginURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/%s/oauth2/v2.0/token", directoryID),
	})

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL.String(), strings.NewReader(url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {defaultLoginScope},
		"grant_type":    {defaultLoginGrantType},
	}.Encode()))
	if err != nil {
		return nil, fmt.Errorf("azure: error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := p.cfg.getHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("azure: error querying oauth2 token: %s", res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("azure: error decoding oauth2 token: %w", err)
	}
	p.token = token

	return p.token, nil
}
