package okta

import (
	"net/http"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/httputil"
)

const (
	// Okta use ISO-8601, see https://developer.okta.com/docs/reference/api-overview/#media-types
	filterDateFormat = "2006-01-02T15:04:05.000Z"

	batchSize        = 200
	readLimit        = 100 * 1024
	httpSuccessClass = 2
)

type config struct {
	apiKey      string
	batchSize   int
	httpClient  *http.Client
	logger      zerolog.Logger
	oktaOptions []okta.ConfigSetter
	url         string
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

// WithOktaOptions sets the okta options in the config.
func WithOktaOptions(oktaOptions ...okta.ConfigSetter) Option {
	return func(cfg *config) {
		cfg.oktaOptions = oktaOptions
	}
}

// WithURL sets the URL option.
func WithURL(url string) Option {
	return func(cfg *config) {
		cfg.url = url
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
