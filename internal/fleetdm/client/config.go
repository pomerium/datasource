package client

import "net/http"

type config struct {
	token               string
	url                 string
	httpClient          *http.Client
	withPolicies        bool
	withVulnerabilities bool
}

type Option func(*config)

var defaults = []Option{
	WithHTTPClient(http.DefaultClient),
}

func newConfig(opts ...Option) *config {
	cfg := new(config)
	for _, opt := range defaults {
		opt(cfg)
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithToken sets the token on the config.
func WithToken(token string) Option {
	return func(cfg *config) {
		cfg.token = token
	}
}

// WithURL sets the URL on the config.
func WithURL(url string) Option {
	return func(cfg *config) {
		cfg.url = url
	}
}

// WithHTTPClient sets the HTTP client on the config.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithPolicies will fetch policy data and populate policy passing fields for hosts.
func WithPolicies() Option {
	return func(cfg *config) {
		cfg.withPolicies = true
	}
}

// WithVulnerabilities will fetch vulnerability data and populate CVE fields for hosts.
func WithVulnerabilities() Option {
	return func(cfg *config) {
		cfg.withVulnerabilities = true
	}
}
