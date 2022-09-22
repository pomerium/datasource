package auth0

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/auth0.v5/management"
)

type (
	// A RoleManager manages roles.
	RoleManager interface {
		List(opts ...management.RequestOption) (r *management.RoleList, err error)
		Users(id string, opts ...management.RequestOption) (u *management.UserList, err error)
	}

	// A UserManager manages users.
	UserManager interface {
		Read(id string, opts ...management.RequestOption) (*management.User, error)
		Roles(id string, opts ...management.RequestOption) (r *management.RoleList, err error)
	}

	newManagersFunc = func(
		ctx context.Context,
		cfg *config,
	) (RoleManager, UserManager, error)
)

func defaultNewManagersFunc(
	ctx context.Context,
	cfg *config,
) (RoleManager, UserManager, error) {
	domain := cfg.domain
	if domain == "" {
		return nil, nil, ErrDomainRequired
	}
	clientID := cfg.clientID
	if clientID == "" {
		return nil, nil, ErrClientIDRequired
	}
	clientSecret := cfg.clientSecret
	if clientSecret == "" {
		return nil, nil, ErrClientSecretRequired
	}

	m, err := management.New(domain,
		management.WithClient(cfg.httpClient),
		management.WithClientCredentials(clientID, clientSecret),
		management.WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: could not build management: %w", err)
	}
	return m.Role, m.User, nil
}

type config struct {
	clientID     string
	clientSecret string
	domain       string
	httpClient   *http.Client
	newManagers  newManagersFunc
}

// Option provides config for the Auth0 Provider.
type Option func(cfg *config)

// WithClientID sets the client id in the config.
func WithClientID(clientID string) Option {
	return func(cfg *config) {
		cfg.clientID = clientID
	}
}

// WithClientSecret sets the client secret in the config.
func WithClientSecret(clientSecret string) Option {
	return func(cfg *config) {
		cfg.clientSecret = clientSecret
	}
}

// WithDomain sets the domain in the config.
func WithDomain(domain string) Option {
	return func(cfg *config) {
		cfg.domain = domain
	}
}

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

func withNewManagersFunc(f newManagersFunc) Option {
	return func(cfg *config) {
		cfg.newManagers = f
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithHTTPClient(http.DefaultClient)
	withNewManagersFunc(defaultNewManagersFunc)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
