package gitlab

import (
	"net/http"
	"net/url"
)

var defaultURL = &url.URL{
	Scheme: "https",
	Host:   "gitlab.com",
}

type config struct {
	httpClient   *http.Client
	privateToken string
	url          *url.URL
}

// An Option updates the gitlab configuration.
type Option func(cfg *config)

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithPrivateToken sets the private token in the config.
func WithPrivateToken(privateToken string) Option {
	return func(cfg *config) {
		cfg.privateToken = privateToken
	}
}

// WithURL sets the api url in the config.
func WithURL(u *url.URL) Option {
	return func(cfg *config) {
		cfg.url = u
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithURL(defaultURL)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
