package onelogin

import (
	"net/http"
	"net/url"
)

type config struct {
	apiURL       *url.URL
	batchSize    int
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// An Option updates the onelogin configuration.
type Option func(cfg *config)

// WithBatchSize sets the batch size option.
func WithBatchSize(batchSize int) Option {
	return func(cfg *config) {
		cfg.batchSize = batchSize
	}
}

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

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithURL sets the api url in the config.
func WithURL(apiURL *url.URL) Option {
	return func(cfg *config) {
		cfg.apiURL = apiURL
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithBatchSize(20)(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithURL(&url.URL{
		Scheme: "https",
		Host:   "api.us.onelogin.com",
	})(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
