package fleetdm

type config struct {
	apiToken           string
	apiURL             string
	certificateQueryID uint
}

type Option func(*config)

var defaults = []Option{}

func newConfig(opts ...Option) *config {
	cfg := new(config)
	for _, opt := range defaults {
		opt(cfg)
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithAPIToken sets the API token on the config.
func WithAPIToken(token string) Option {
	return func(cfg *config) {
		cfg.apiToken = token
	}
}

// WithAPIURL sets the API URL on the config.
func WithAPIURL(url string) Option {
	return func(cfg *config) {
		cfg.apiURL = url
	}
}

// WithCertificateQueryID sets the certificate query ID on the config.
func WithCertificateQueryID(id uint) Option {
	return func(cfg *config) {
		cfg.certificateQueryID = id
	}
}
