package azure

import (
	"net/http"
	"net/url"
)

const (
	defaultGraphHost = "graph.microsoft.com"

	defaultLoginHost      = "login.microsoftonline.com"
	defaultLoginScope     = "https://graph.microsoft.com/.default"
	defaultLoginGrantType = "client_credentials"
)

type config struct {
	clientID     string
	clientSecret string
	directoryID  string
	graphURL     *url.URL
	httpClient   *http.Client
	loginURL     *url.URL
}

// An Option updates the provider configuration.
type Option func(cfg *config)

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

// WithDirectoryID sets the directory in the config.
func WithDirectoryID(directoryID string) Option {
	return func(cfg *config) {
		cfg.directoryID = directoryID
	}
}

// WithGraphURL sets the graph URL for the configuration.
func WithGraphURL(graphURL *url.URL) Option {
	return func(cfg *config) {
		cfg.graphURL = graphURL
	}
}

// WithHTTPClient sets the http client to use for requests to the Azure APIs.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithLoginURL sets the login URL for the configuration.
func WithLoginURL(loginURL *url.URL) Option {
	return func(cfg *config) {
		cfg.loginURL = loginURL
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithGraphURL(&url.URL{
		Scheme: "https",
		Host:   defaultGraphHost,
	})(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithLoginURL(&url.URL{
		Scheme: "https",
		Host:   defaultLoginHost,
	})(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
