package keycloak

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/httputil"
)

const (
	// DefaultBatchSize is the default batch size for querying groups and users.
	DefaultBatchSize = 100
	// DefaultRealm is the default realm if one is not provided.
	DefaultRealm = "master"
)

type config struct {
	batchSize    int
	httpClient   *http.Client
	clientID     string
	clientSecret string
	logger       zerolog.Logger
	realm        string
	url          string
}

// An Option configures the Keycloak Provider.
type Option func(cfg *config)

// WithBatchSize sets the batch size in the config.
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

// WithRealm sets the realm in the config.
func WithRealm(realm string) Option {
	return func(cfg *config) {
		cfg.realm = realm
	}
}

// WithHTTPClient sets the http client in the config.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithLogger sets the logger in the config.
func WithLogger(logger zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// WithURL sets the url in the config. URL should be the base url without /realms/master.
func WithURL(url string) Option {
	return func(cfg *config) {
		cfg.url = url
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithBatchSize(DefaultBatchSize)(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithLogger(log.Logger)(cfg)
	WithRealm(DefaultRealm)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

func (cfg *config) getHTTPClient() *http.Client {
	return httputil.NewLoggingClient(cfg.logger, cfg.httpClient, func(event *zerolog.Event) *zerolog.Event {
		return event.Str("idp", "keycloak")
	})
}
