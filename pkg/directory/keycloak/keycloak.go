// Package keycloak contains a Keycloak directory provider.
package keycloak

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pomerium/datasource/internal/jsonutil"
)

type apiGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}

func listGroups(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	realm string,
	batchSize int,
) iter.Seq2[apiGroup, error] {
	apiURL := joinURL(baseURL, "/admin/realms/"+url.PathEscape(realm)+"/groups")
	return list[apiGroup](ctx, client, apiURL, batchSize)
}

func listGroupMembers(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	realm string,
	groupID string,
	batchSize int,
) iter.Seq2[apiUser, error] {
	apiURL := joinURL(baseURL, "/admin/realms/"+url.PathEscape(realm)+"/groups/"+url.PathEscape(groupID)+"/members")
	return list[apiUser](ctx, client, apiURL, batchSize)
}

func listUsers(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	realm string,
	batchSize int,
) iter.Seq2[apiUser, error] {
	apiURL := joinURL(baseURL, "/admin/realms/"+url.PathEscape(realm)+"/users")
	return list[apiUser](ctx, client, apiURL, batchSize)
}

func list[T any](
	ctx context.Context,
	client *http.Client,
	apiURL string,
	batchSize int,
) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var def T

		apiURL, err := url.Parse(apiURL)
		if err != nil {
			yield(def, err)
			return
		}

		for i := 0; ; i += batchSize {
			apiURL = apiURL.ResolveReference(&url.URL{
				RawQuery: url.Values{
					"first": {strconv.Itoa(i)},
					"max":   {strconv.Itoa(batchSize)},
				}.Encode(),
			})

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
			if err != nil {
				yield(def, fmt.Errorf("error creating api http request: %w", err))
				return
			}
			req.Header.Set("Accept", "application/json")

			res, err := client.Do(req)
			if err != nil {
				yield(def, fmt.Errorf("error making api http request: %w", err))
				return
			}
			defer res.Body.Close()

			if res.StatusCode/100 != 2 {
				yield(def, fmt.Errorf("error listing, unexpected status code %d", res.StatusCode))
				return
			}

			cnt := 0
			for v, err := range jsonutil.StreamArrayReader[T](res.Body, nil) {
				if err != nil {
					yield(v, err)
					return
				}

				if !yield(v, err) {
					return
				}

				cnt++
			}

			if cnt < batchSize {
				break
			}
		}
	}
}
