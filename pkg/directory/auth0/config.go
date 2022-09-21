package auth0

import (
	"context"
	"fmt"

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

	newManagersFunc = func(ctx context.Context, domain string, serviceAccount *ServiceAccount) (RoleManager, UserManager, error)
)

func defaultNewManagersFunc(ctx context.Context, domain string, serviceAccount *ServiceAccount) (RoleManager, UserManager, error) {
	// override the domain for the management api if supplied
	if serviceAccount.Domain != "" {
		domain = serviceAccount.Domain
	}
	m, err := management.New(domain,
		management.WithClientCredentials(serviceAccount.ClientID, serviceAccount.Secret),
		management.WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("auth0: could not build management: %w", err)
	}
	return m.Role, m.User, nil
}

type config struct {
	domain         string
	serviceAccount *ServiceAccount
	newManagers    newManagersFunc
}

// Option provides config for the Auth0 Provider.
type Option func(cfg *config)

// WithDomain sets the provider domain option.
func WithDomain(domain string) Option {
	return func(cfg *config) {
		cfg.domain = domain
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
	withNewManagersFunc(defaultNewManagersFunc)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
