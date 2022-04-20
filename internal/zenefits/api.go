package zenefits

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Auth is API server auth parameters
type Auth struct {
	// Base is API URL, nil for default
	BaseURL *url.URL `validate:"-"`
}

// RequestURL returns request URL for a given core object
// by default, there's just one company corresponding to the api token
// however, according to API there may be multiple, in that case
// base API need be initialized to "/core/companies/{company_id}/"
func (a Auth) RequestURL(rel string) *url.URL {
	base := a.BaseURL
	if base == nil {
		base = &url.URL{
			Scheme: "https",
			Host:   "api.zenefits.com",
			Path:   "/core/",
		}
	}
	u := base.ResolveReference(&url.URL{Path: rel})
	return u
}

type PeopleRequest struct {
	Auth
	DepartmentID *string
	Status       *string
	LocationID   *string
	Fields       []string
}

func (req *PeopleRequest) getURL() string {
	u := req.Auth.RequestURL("people")

	param := make(url.Values)
	param.Set("includes", "department location")
	if req.DepartmentID != nil {
		param.Set("department", fmt.Sprint(*req.DepartmentID))
	}
	if req.Status != nil {
		param.Set("status", fmt.Sprint(*req.Status))
	}
	if req.LocationID != nil {
		param.Set("location", fmt.Sprint(*req.LocationID))
	}

	u.RawQuery = param.Encode()

	return u.String()
}

func (req *PeopleRequest) GetRequest(ctx context.Context) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, req.getURL(), http.NoBody)
	if err != nil {
		return nil, err
	}
	r.Header.Add("accept", "application/json")
	return r, nil
}

// GetEmployees returns all employees for a company
// this method may be optimized by avoiding referenced objects inlining
// and fetching departments and locations (and other referenced fields, if necessary)
// individually or via list
func GetEmployees(ctx context.Context, client *http.Client, param PeopleRequest) ([]Person, error) {
	req, err := param.GetRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	res, err := doListRequest[Person](client, req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	return res, nil
}

func doListRequest[T Kind](client *http.Client, req *http.Request) ([]T, error) {
	ctx := req.Context()
	var items []T

	for {
		resp, err := doOneRequest[T](client, req)
		if err != nil {
			return nil, err
		}
		items = append(items, resp.Data.Items...)

		if resp.Data.NextURL == nil {
			break
		}

		req = req.Clone(ctx)
		req.URL = resp.Data.NextURL.URL()
	}

	return items, nil
}

func doOneRequest[T Kind](client *http.Client, req *http.Request) (*Response[T], error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %v: unexpected return status: %s", req.URL.String(), resp.Status)
	}

	dst := new(Response[T])
	if err = json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return nil, fmt.Errorf("decode json response: %w", err)
	}

	return dst, nil
}
