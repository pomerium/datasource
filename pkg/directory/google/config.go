package google

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultProviderURL = "https://www.googleapis.com/"
)

type config struct {
	logger          zerolog.Logger
	impersonateUser string
	jsonKey         []byte
	jsonKeyFile     string
	url             string
}

// An Option changes the configuration for the Google directory provider.
type Option func(cfg *config)

// WithImpersonateUser sets the impersonate user in the config.
func WithImpersonateUser(impersonateUser string) Option {
	return func(cfg *config) {
		cfg.impersonateUser = impersonateUser
	}
}

// WithJSONKey sets the json key in the config.
func WithJSONKey(jsonKey []byte) Option {
	return func(cfg *config) {
		cfg.jsonKey = jsonKey
	}
}

// WithJSONKeyFile sets the json key file in the config.
func WithJSONKeyFile(jsonKeyFile string) Option {
	return func(cfg *config) {
		cfg.jsonKeyFile = jsonKeyFile
	}
}

// WithLogger sets the logger in the config.
func WithLogger(logger zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
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
	WithLogger(log.Logger)(cfg)
	WithURL(defaultProviderURL)(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func (cfg *config) getJSONKey() ([]byte, error) {
	if cfg.jsonKeyFile != "" {
		return os.ReadFile(cfg.jsonKeyFile)
	}
	return cfg.jsonKey, nil
}
