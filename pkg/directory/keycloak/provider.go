package keycloak

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/pomerium/datasource/pkg/directory"
)

// A Provider implements the directory provider interface for Keycloak.
type Provider struct {
	cfg *config

	tokenSourceMu sync.Mutex
	tokenSource   oauth2.TokenSource
}

// New creates a new provider.
func New(options ...Option) *Provider {
	return &Provider{
		cfg: getConfig(options...),
	}
}

// GetDirectory returns the directory information for a Keycloak realm.
func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	client, err := p.getHTTPClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	var dgs []directory.Group
	for g, err := range listGroups(ctx, client, p.cfg.url, p.cfg.realm, p.cfg.batchSize) {
		if err != nil {
			return nil, nil, err
		}

		dgs = append(dgs, directory.Group{
			ID:   g.ID,
			Name: g.Name,
		})
	}
	slices.SortFunc(dgs, func(dg1, dg2 directory.Group) int {
		return cmp.Compare(dg1.ID, dg2.ID)
	})

	groupLookup := map[string][]string{}
	for _, dg := range dgs {
		for u, err := range listGroupMembers(ctx, client, p.cfg.url, p.cfg.realm, dg.ID, p.cfg.batchSize) {
			if err != nil {
				return nil, nil, err
			}
			groupLookup[u.ID] = append(groupLookup[u.ID], dg.ID)
		}
	}

	var dus []directory.User
	for u, err := range listUsers(ctx, client, p.cfg.url, p.cfg.realm, p.cfg.batchSize) {
		if err != nil {
			return nil, nil, err
		}

		email := u.Email
		if !u.EmailVerified {
			email = ""
		}
		dus = append(dus, directory.User{
			ID:          u.ID,
			DisplayName: u.Username,
			GroupIDs:    groupLookup[u.ID],
			Email:       email,
		})
	}
	slices.SortFunc(dus, func(du1, du2 directory.User) int {
		return cmp.Compare(du1.ID, du2.ID)
	})

	return dgs, dus, nil
}

func (p *Provider) getHTTPClient(ctx context.Context) (*http.Client, error) {
	p.tokenSourceMu.Lock()
	defer p.tokenSourceMu.Unlock()

	if p.cfg.realm == "" {
		return nil, fmt.Errorf("realm is required")
	}
	if p.cfg.clientID == "" {
		return nil, fmt.Errorf("client id is required")
	}
	if p.cfg.clientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}
	if p.cfg.url == "" {
		return nil, fmt.Errorf("url is required")
	}

	client := p.cfg.getHTTPClient()

	// set up the token source for oauth
	if p.tokenSource == nil {
		tokenSourceCtx := context.Background()
		tokenSourceCtx = oidc.ClientContext(tokenSourceCtx, client)

		oidcProvider, err := oidc.NewProvider(tokenSourceCtx,
			joinURL(p.cfg.url, "/realms/"+url.PathEscape(p.cfg.realm)))
		if err != nil {
			return nil, fmt.Errorf("error creating oidc provider (url=%s): %w", p.cfg.url, err)
		}

		e := oidcProvider.Endpoint()
		p.tokenSource = (&clientcredentials.Config{
			ClientID:     p.cfg.clientID,
			ClientSecret: p.cfg.clientSecret,
			TokenURL:     e.TokenURL,
			AuthStyle:    e.AuthStyle,
		}).TokenSource(tokenSourceCtx)
	}

	return oauth2.NewClient(oidc.ClientContext(ctx, client), p.tokenSource), nil
}

func joinURL(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}
