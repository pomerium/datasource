package google

import "net/http"

const (
	defaultProviderURL = "https://www.googleapis.com/"
)

type config struct {
	httpClient     *http.Client
	serviceAccount *ServiceAccount
	url            string
}

// An Option changes the configuration for the Google directory provider.
type Option func(cfg *config)

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithServiceAccount sets the service account in the Google configuration.
func WithServiceAccount(serviceAccount *ServiceAccount) Option {
	return func(cfg *config) {
		cfg.serviceAccount = serviceAccount
	}
}

// WithURL sets the provider url to use.
func WithURL(url string) Option {
	return func(cfg *config) {
		cfg.url = url
	}
}

func getConfig(opts ...Option) *config {
	cfg := new(config)
	WithURL(defaultProviderURL)(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
