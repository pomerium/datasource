package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog"
)

type debug struct {
	http.RoundTripper
	zerolog.Logger
}

// RoundTrip dumps request/response
func (t *debug) RoundTrip(req *http.Request) (*http.Response, error) {
	data, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		t.Err(err).Msg("dump request")
		return nil, err
	}
	fmt.Println(string(data))

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		t.Err(err).Msg("request error")
		return nil, err
	}
	data, err = httputil.DumpResponse(resp, true)
	if err != nil {
		t.Err(err).Msg("dump response")
		return resp, nil
	}
	fmt.Println(string(data))
	return resp, nil
}

// NewDebugClient creates a round tripper that
func NewDebugClient(base *http.Client, log zerolog.Logger) *http.Client {
	client := *base
	tr := &debug{client.Transport, log}
	if tr.RoundTripper == nil {
		tr.RoundTripper = http.DefaultTransport
	}
	client.Transport = tr
	return &client
}
