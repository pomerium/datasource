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

	newManagersFunc = func(ctx context.Context, httpClient *http.Client, serviceAccount *ServiceAccount) (RoleManager, UserManager, error)
)

func defaultNewManagersFunc(ctx context.Context, httpClient *http.Client, serviceAccount *ServiceAccount) (RoleManager, UserManager, error) {
	m, err := management.New(serviceAccount.Domain,
		management.WithClient(httpClient),
		management.WithClientCredentials(serviceAccount.ClientID, serviceAccount.ClientSecret),
		management.WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: could not build management: %w", err)
	}
	return m.Role, m.User, nil
}

type config struct {
	serviceAccount *ServiceAccount
	httpClient     *http.Client
	newManagers    newManagersFunc
}

// Option provides config for the Auth0 Provider.
type Option func(cfg *config)

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithServiceAccount sets the service account option.
func WithServiceAccount(serviceAccount *ServiceAccount) Option {
	return func(cfg *config) {
		cfg.serviceAccount = serviceAccount
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
