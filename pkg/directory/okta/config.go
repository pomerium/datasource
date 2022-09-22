package okta

import (
	"net/http"
	"net/url"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/httputil"
)

const (
	// Okta use ISO-8601, see https://developer.okta.com/docs/reference/api-overview/#media-types
	filterDateFormat = "2006-01-02T15:04:05.999Z"

	batchSize        = 200
	readLimit        = 100 * 1024
	httpSuccessClass = 2
)

type config struct {
	apiKey      string
	batchSize   int
	httpClient  *http.Client
	logger      zerolog.Logger
	providerURL *url.URL
}

// An Option configures the Okta Provider.
type Option func(cfg *config)

// WithAPIKey sets the api key in the config.
func WithAPIKey(apiKey string) Option {
	return func(cfg *config) {
		cfg.apiKey = apiKey
	}
}

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

// WithLogger sets the logger in the config.
func WithLogger(logger zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// WithProviderURL sets the provider URL option.
func WithProviderURL(uri *url.URL) Option {
	return func(cfg *config) {
		cfg.providerURL = uri
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithBatchSize(batchSize)(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithLogger(log.Logger)(cfg)
	for _, option := range options {
		option(cfg)
	}

	return cfg
}

func (cfg *config) getHTTPClient() *http.Client {
	return httputil.NewLoggingClient(cfg.logger, cfg.httpClient, func(event *zerolog.Event) *zerolog.Event {
		return event.Str("idp", "okta")
	})
}
