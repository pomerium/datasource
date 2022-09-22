package httputil

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggingRoundTripper struct {
	base      http.RoundTripper
	logger    zerolog.Logger
	customize []func(event *zerolog.Event) *zerolog.Event
}

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	res, err := l.base.RoundTrip(req)
	statusCode := http.StatusInternalServerError
	if res != nil {
		statusCode = res.StatusCode
	}
	evt := l.logger.Debug().
		Str("method", req.Method).
		Str("authority", req.URL.Host).
		Str("path", req.URL.Path).
		Dur("duration", time.Since(start)).
		Int("response-code", statusCode)
	for _, f := range l.customize {
		f(evt)
	}
	evt.Msg("http-request")
	return res, err
}

// NewLoggingRoundTripper creates a http.RoundTripper that will log requests.
func NewLoggingRoundTripper(logger zerolog.Logger, base http.RoundTripper, customize ...func(event *zerolog.Event) *zerolog.Event) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return loggingRoundTripper{base: base, logger: logger, customize: customize}
}

// NewLoggingClient creates a new http.Client that will log requests.
func NewLoggingClient(logger zerolog.Logger, base *http.Client, customize ...func(event *zerolog.Event) *zerolog.Event) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	newClient := new(http.Client)
	*newClient = *base
	newClient.Transport = NewLoggingRoundTripper(logger, newClient.Transport, customize...)
	return newClient
}
