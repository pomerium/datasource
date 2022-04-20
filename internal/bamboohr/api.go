package bamboohr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/pomerium/datasource/internal/util"
)

// Auth is API server auth parameters
type Auth struct {
	// APIKey see https://documentation.bamboohr.com/docs#section-authentication
	APIKey string `validate:"required"`
	// Subdomain is BambooHR customer identifier
	// If you access BambooHR at https://mycompany.bamboohr.com, then the subdomain is "mycompany"
	Subdomain string `validate:"required"`
	// Base is API URL, nil for default
	BaseURL *url.URL
}

func (a Auth) RequestURL(rel string) *url.URL {
	base := a.BaseURL
	if base == nil {
		base = &url.URL{
			Scheme: "https",
			Host:   "api.bamboohr.com",
			Path:   "/api/gateway.php/",
		}
	}
	u := *base
	u.Path = path.Join(base.Path, a.Subdomain, rel)
	u.User = url.UserPassword(a.APIKey, "x")
	return &u
}

// EmployeeRequest requests
type EmployeeRequest struct {
	Auth
	// CurrentOnly returns only currently active employees
	CurrentOnly bool
	// Fields specifies fields that should be returned
	// https://documentation.bamboohr.com/docs/list-of-field-names
	// https://documentation.bamboohr.com/docs/field-types
	Fields []string
	// Remap instructs to remap certain fields in each entry
	Remap []util.FieldRemap
}

func (req EmployeeRequest) RequestURL() *url.URL {
	u := req.Auth.RequestURL("v1/reports/custom")

	vals := make(url.Values)
	vals.Add("format", "json")
	vals.Add("onlyCurrent", fmt.Sprint(req.CurrentOnly))
	u.RawQuery = vals.Encode()

	return u
}

// EmployeeDirectory returns full list of employees
func GetEmployees(ctx context.Context, client *http.Client, param EmployeeRequest) ([]map[string]interface{}, error) {
	body, err := getEmployeesRequestBody(param.Fields)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	u := param.RequestURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST %v: unexpected return status: %s", u, resp.Status)
	}

	employees, err := parseEmployeesResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get employees: %w", err)
	}

	if param.Remap == nil {
		return employees, nil
	}

	if err = util.Remap(employees, param.Remap); err != nil {
		return nil, fmt.Errorf("remap employee fields: %w", err)
	}

	return employees, nil
}

func getEmployeesRequestBody(fields []string) (io.ReadCloser, error) {
	var buf bytes.Buffer
	req := struct {
		Fields []string `json:"fields"`
	}{fields}
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}
	return io.NopCloser(&buf), nil
}

func parseEmployeesResponse(src io.Reader) ([]map[string]interface{}, error) {
	var dst struct {
		Employees []map[string]interface{} `json:"employees"`
	}
	if err := json.NewDecoder(src).Decode(&dst); err != nil {
		return nil, err
	}
	return dst.Employees, nil
}
