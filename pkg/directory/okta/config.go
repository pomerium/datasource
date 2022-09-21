package okta

import (
	"net/http"
	"net/url"
)

const (
	// Okta use ISO-8601, see https://developer.okta.com/docs/reference/api-overview/#media-types
	filterDateFormat = "2006-01-02T15:04:05.999Z"

	batchSize        = 200
	readLimit        = 100 * 1024
	httpSuccessClass = 2
)

type config struct {
	batchSize      int
	httpClient     *http.Client
	providerURL    *url.URL
	serviceAccount *ServiceAccount
}

// An Option configures the Okta Provider.
type Option func(cfg *config)

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

// WithProviderURL sets the provider URL option.
func WithProviderURL(uri *url.URL) Option {
	return func(cfg *config) {
		cfg.providerURL = uri
	}
}

// WithServiceAccount sets the service account option.
func WithServiceAccount(serviceAccount *ServiceAccount) Option {
	return func(cfg *config) {
		cfg.serviceAccount = serviceAccount
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithBatchSize(batchSize)(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	for _, option := range options {
		option(cfg)
	}

	return cfg
}
