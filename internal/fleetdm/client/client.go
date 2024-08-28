package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.token))
	return req.WithContext(ctx), nil
}

func (c *Client) ListHosts(
	ctx context.Context,
) ([]Host, error) {
	req, err := c.newRequest(ctx, "GET", "/api/v1/fleet/hosts")
	if err != nil {
		return nil, err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Hosts []Host `json:"hosts"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Hosts, nil
}

func (c *Client) QueryReport(
	ctx context.Context,
	queryID uint,
) ([]QueryReportItem, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/api/v1/fleet/queries/%d/report", queryID))
	if err != nil {
		return nil, err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Results []QueryReportItem `json:"results"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Results, nil
}
