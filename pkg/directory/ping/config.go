package ping

import (
	"net/http"
	"net/url"
	"strings"
)

type config struct {
	authURL       *url.URL
	apiURL        *url.URL
	clientID      string
	clientSecret  string
	httpClient    *http.Client
	environmentID string
}

// An Option updates the Ping configuration.
type Option func(cfg *config)

// WithAPIURL sets the api url in the config.
func WithAPIURL(apiURL *url.URL) Option {
	return func(cfg *config) {
		cfg.apiURL = apiURL
	}
}

// WithAuthURL sets the auth url in the config.
func WithAuthURL(authURL *url.URL) Option {
	return func(cfg *config) {
		cfg.authURL = authURL
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

// WithEnvironmentID sets the environment ID in the config.
func WithEnvironmentID(environmentID string) Option {
	return func(cfg *config) {
		cfg.environmentID = environmentID
	}
}

// WithHTTPClient sets the http client option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithProviderURL sets the environment ID from the provider URL set in the config.
func WithProviderURL(providerURL *url.URL) Option {
	// provider URL will be https://auth.pingone.com/{ENVIRONMENT_ID}/as
	if providerURL == nil {
		return func(cfg *config) {}
	}
	parts := strings.Split(providerURL.Path, "/")
	if len(parts) < 1 {
		return func(cfg *config) {}
	}
	return WithEnvironmentID(parts[1])
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithAuthURL(&url.URL{
		Scheme: "https",
		Host:   "auth.pingone.com",
	})(cfg)
	WithAPIURL(&url.URL{
		Scheme: "https",
		Host:   "api.pingone.com",
	})(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
