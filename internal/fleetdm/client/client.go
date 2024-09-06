package client

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"github.com/pomerium/datasource/internal/jsonutil"
)

type Client struct {
	cfg *config
}

// New creates a new FleetDM API client
// see https://fleetdm.com/docs/rest-api/rest-api
func New(opts ...Option) (*Client, error) {
	cfg := newConfig(opts...)
	return &Client{
		cfg: cfg,
	}, nil
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
) (*http.Request, error) {
	u, err := url.Parse(c.cfg.url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse api endpoint URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("api endpoint URL scheme must be http or https")
	}
	u.Path = path

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.token))
	return req.WithContext(ctx), nil
}

func (c *Client) ListHosts(
	ctx context.Context,
) (iter.Seq2[Host, error], error) {
	req, err := c.newRequest(ctx, "GET", "/api/v1/fleet/hosts")
	if err != nil {
		return nil, err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return convertIter2(
		jsonutil.StreamArrayReadAndClose[hostRecord](resp.Body, []string{"hosts"}),
		convertHostRecord,
	), nil
}

func (c *Client) QueryCertificates(
	ctx context.Context,
	queryID uint,
) (iter.Seq2[CertificateSHA1QueryItem, error], error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/api/v1/fleet/queries/%d/report", queryID))
	if err != nil {
		return nil, err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return convertIter2(
		jsonutil.StreamArrayReadAndClose[certificateQueryRecord](resp.Body, []string{"results"}),
		convertCertificateQuery,
	), nil
}
