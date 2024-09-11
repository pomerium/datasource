package client

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"github.com/pomerium/datasource/internal/jsonutil"

	"github.com/hashicorp/go-set/v3"
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

func (c *Client) ListHosts(
	ctx context.Context,
) (iter.Seq2[Host, error], error) {
	return fetchItems(ctx, c, convertHostRecord, "hosts", "/api/v1/fleet/hosts", "populate_software", "true", "populate_policies", "true")
}

func (c *Client) listTeams(ctx context.Context) ([]uint, error) {
	iter, err := fetchItems(ctx, c,
		func(tm struct {
			ID uint `json:"id"`
		},
		) (uint, error) {
			return tm.ID, nil
		},
		"teams", "/api/v1/fleet/teams")
	if err != nil {
		return nil, err
	}

	var ids []uint
	for id, err := range iter {
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (c *Client) ListPolicies(ctx context.Context) (iter.Seq2[Policy, error], error) {
	teams, err := c.listTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	global, err := fetchItems(ctx, c, convertPolicy, "policies", "/api/latest/fleet/policies")
	if err != nil {
		return nil, fmt.Errorf("list global policies: %w", err)
	}

	policies := []iter.Seq2[Policy, error]{global}
	for _, teamID := range teams {
		p, err := fetchItems(ctx, c, convertPolicy, "policies", fmt.Sprintf("/api/latest/fleet/teams/%d/policies", teamID))
		if err != nil {
			return nil, fmt.Errorf("list team policies: %w", err)
		}
		policies = append(policies, p)
	}

	return dedup(policies...), nil
}

func (c *Client) QueryCertificates(
	ctx context.Context,
	queryID uint,
) (iter.Seq2[CertificateSHA1QueryItem, error], error) {
	return fetchItems(ctx, c, convertCertificateQuery, "results", fmt.Sprintf("/api/v1/fleet/queries/%d/report", queryID))
}

func fetchItems[InternalRecord, ExternalRecord any](
	ctx context.Context,
	c *Client,
	convert func(InternalRecord) (ExternalRecord, error),
	key string,
	path string,
	args ...string,
) (iter.Seq2[ExternalRecord, error], error) {
	req, err := c.newRequest(ctx, "GET", path, args...)
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
		jsonutil.StreamArrayReadAndClose[InternalRecord](resp.Body, []string{key}),
		convert,
	), nil
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
	kv ...string,
) (*http.Request, error) {
	u, err := url.Parse(c.cfg.url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse api endpoint URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("api endpoint URL scheme must be http or https")
	}
	u.Path = path

	if len(kv)%2 != 0 {
		return nil, fmt.Errorf("key-value pairs must be even")
	}

	query := make(url.Values)
	for i := 0; i < len(kv); i += 2 {
		query.Add(kv[i], kv[i+1])
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.token))
	return req.WithContext(ctx), nil
}

func dedup[ID comparable, T interface{ GetID() ID }](
	iters ...iter.Seq2[T, error],
) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		seen := set.New[ID](0)
		for _, iter := range iters {
			for v, err := range iter {
				if err != nil {
					if !yield(v, err) {
						return
					}
					continue
				}
				id := v.GetID()
				if seen.Contains(id) {
					continue
				}
				seen.Insert(id)
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}
