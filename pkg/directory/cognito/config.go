package cognito

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/httputil"
)

type config struct {
	accessKeyID     string
	httpClient      *http.Client
	logger          zerolog.Logger
	secretAccessKey string
	sessionToken    string
	userPoolID      string
}

type Option func(cfg *config)

// WithAccessKeyID sets the access key id config option.
func WithAccessKeyID(accessKeyID string) Option {
	return func(cfg *config) {
		cfg.accessKeyID = accessKeyID
	}
}

// WithHTTPClient sets the http client config option.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = httpClient
	}
}

// WithLogger sets the logger config option.
func WithLogger(logger zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// WithSecretAccessKey sets the secret access key config option.
func WithSecretAccessKey(secretAccessKey string) Option {
	return func(cfg *config) {
		cfg.secretAccessKey = secretAccessKey
	}
}

// WithSessionToken sets the session token config option.
func WithSessionToken(sessionToken string) Option {
	return func(cfg *config) {
		cfg.sessionToken = sessionToken
	}
}

// WithUserPoolID sets the user pool ID config option.
func WithUserPoolID(userPoolID string) Option {
	return func(cfg *config) {
		cfg.userPoolID = userPoolID
	}
}

func getConfig(options ...Option) *config {
	cfg := new(config)
	WithHTTPClient(http.DefaultClient)(cfg)
	WithLogger(log.Logger)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

func (cfg *config) getHTTPClient() *http.Client {
	return httputil.NewLoggingClient(cfg.logger, cfg.httpClient, func(event *zerolog.Event) *zerolog.Event {
		return event.Str("idp", "cognito")
	})
}
