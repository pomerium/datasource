package onelogin

import (
	"net/http"
	"net/url"
)

type config struct {
	apiURL         *url.URL
	batchSize      int
	serviceAccount *ServiceAccount
	httpClient     *http.Client
}

// An Option updates the onelogin configuration.
type Option func(*config)

// WithBatchSize sets the batch size option.
func WithBatchSize(batchSize int) Option {
	return func(cfg *config) {
		cfg.batchSize = batchSize
	}
}

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithServiceAccount sets the service account in the config.
func WithServiceAccount(serviceAccount *ServiceAccount) Option {
	return func(cfg *config) {
		cfg.serviceAccount = serviceAccount
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
