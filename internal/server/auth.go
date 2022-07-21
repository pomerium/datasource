package server

import (
	"fmt"
	"net/http"

	"github.com/pomerium/datasource/internal/version"
)

type transport struct {
	token string
	rt    http.RoundTripper
}

// RoundTrip executes a single HTTP transaction
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.ProjectName, version.Version))
	return t.rt.RoundTrip(req)
}

func NewBearerTokenClient(base *http.Client, token string) *http.Client {
	client := *base
	tr := &transport{token, client.Transport}
	if tr.rt == nil {
		tr.rt = http.DefaultTransport
	}
	client.Transport = tr
	return &client
}
