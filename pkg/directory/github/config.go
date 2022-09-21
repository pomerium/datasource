package github

import (
	"net/http"
	"net/url"
)

var defaultURL = &url.URL{
	Scheme: "https",
	Host:   "api.github.com",
}

type config struct {
	httpClient          *http.Client
	personalAccessToken string
	url                 *url.URL
	username            string
}

// An Option updates the github configuration.
type Option func(cfg *config)

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithPersonalAccessToken sets the personal access token in the config.
func WithPersonalAccessToken(personalAccessToken string) Option {
	return func(cfg *config) {
		cfg.personalAccessToken = personalAccessToken
	}
}

// WithURL sets the api url in the config.
func WithURL(u *url.URL) Option {
	return func(cfg *config) {
		cfg.url = u
	}
}

// WithUsername sets the username in the config.
func WithUsername(username string) Option {
	return func(cfg *config) {
		cfg.username = username
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
