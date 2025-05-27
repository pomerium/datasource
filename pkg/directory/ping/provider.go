package ping

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/pomerium/datasource/pkg/directory"
)

// Provider implements a directory provider using the Ping API.
type Provider struct {
	cfg   *config
	mu    sync.RWMutex
	token *oauth2.Token
}

// New creates a new Ping Provider.
func New(options ...Option) *Provider {
	cfg := getConfig(options...)
	return &Provider{
		cfg: cfg,
	}
}

// GetDirectory returns all the users and groups in the directory.
func (p *Provider) GetDirectory(ctx context.Context) (directory.Bundle, error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return directory.Bundle{}, err
	}

	apiGroups, err := getAllGroups(ctx, client, p.cfg.apiURL, p.cfg.environmentID)
	if err != nil {
		return directory.Bundle{}, err
	}

	directoryUserLookup := map[string]directory.User{}
	directoryGroups := make([]directory.Group, len(apiGroups))
	for i, ag := range apiGroups {
		dg := directory.Group{
			ID:   ag.ID,
			Name: ag.Name,
		}

		apiUsers, err := getGroupUsers(ctx, client, p.cfg.apiURL, p.cfg.environmentID, ag.ID)
		if err != nil {
			return directory.Bundle{}, err
		}
		for _, au := range apiUsers {
			du, ok := directoryUserLookup[au.ID]
			if !ok {
				du = directory.User{
					ID:          au.ID,
					DisplayName: au.getDisplayName(),
					Email:       au.Email,
				}
			}
			du.GroupIDs = append(du.GroupIDs, ag.ID)
			directoryUserLookup[au.ID] = du
		}

		directoryGroups[i] = dg
	}
	sort.Slice(directoryGroups, func(i, j int) bool {
		return directoryGroups[i].ID < directoryGroups[j].ID
	})

	directoryUsers := make([]directory.User, 0, len(directoryUserLookup))
	for _, du := range directoryUserLookup {
		directoryUsers = append(directoryUsers, du)
	}
	sort.Slice(directoryUsers, func(i, j int) bool {
		return directoryUsers[i].ID < directoryUsers[j].ID
	})

	return directory.NewBundle(directoryGroups, directoryUsers, nil), nil
}

func (p *Provider) getClient(ctx context.Context) (*http.Client, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, err
	}

	client := p.cfg.getHTTPClient()
	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(token),
		Base:   client.Transport,
	}
	return client, nil
}

func (p *Provider) getToken(ctx context.Context) (*oauth2.Token, error) {
	environmentID := p.cfg.environmentID
	if environmentID == "" {
		return nil, ErrEnvironmentIDRequired
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

	ocfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL: p.cfg.authURL.ResolveReference(&url.URL{
			Path: fmt.Sprintf("/%s/as/token", environmentID),
		}).String(),
	}
	var err error
	p.token, err = ocfg.Token(ctx)
	if err != nil {
		return nil, err
	}

	return p.token, nil
}
