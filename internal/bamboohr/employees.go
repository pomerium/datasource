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
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
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

// EmployeeRequest requests employees
type EmployeeRequest struct {
	Auth
	Location *time.Location
}

func (req EmployeeRequest) RequestURL() *url.URL {
	u := req.Auth.RequestURL("v1/reports/custom")

	vals := make(url.Values)
	vals.Add("format", "json")
	vals.Add("onlyCurrent", "true")
	u.RawQuery = vals.Encode()

	return u
}

// Employee represents BambooHR employee record
type Employee struct {
	ID         json.Number `json:"bamboo_id" mapstructure:"id"`
	Email      string      `json:"id" mapstructure:"workEmail"`
	Department string      `json:"department" mapstructure:"department"`
	Divison    string      `json:"division" mapstructure:"division"`
	Status     string      `json:"status" mapstructure:"status"`
	FirstName  string      `json:"first_name" mapstructure:"firstName"`
	LastName   string      `json:"last_name" mapstructure:"lastName"`
	Country    string      `json:"country" mapstructure:"country"`
	State      string      `json:"state" mapstructure:"state"`
}

var (
	// JSON tags represent how data is produced to the outside consumer
	// mapstructure tags match the internal BambooHR field naming
	employeeRequestFields = util.GetStructTagNames(Employee{}, "mapstructure")
)

// GetAllEmployees returns full list of employees in active status
func GetAllEmployees(ctx context.Context, client *http.Client, param EmployeeRequest) ([]Employee, error) {
	body, err := getEmployeesRequestBody()
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

	return employees, nil
}

func getEmployeesRequestBody() (io.ReadCloser, error) {
	var buf bytes.Buffer
	req := struct {
		Fields []string `json:"fields"`
	}{
		Fields: employeeRequestFields,
	}
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, err
	}
	return io.NopCloser(&buf), nil
}

type field struct {
	ID string `json:"id"`
}

func parseEmployeesResponse(src io.Reader) ([]Employee, error) {
	var dst struct {
		Fields    []field                  `json:"fields"`
		Employees []map[string]interface{} `json:"employees"`
	}
	if err := json.NewDecoder(src).Decode(&dst); err != nil {
		return nil, err
	}

	if err := checkFieldsPresent(employeeRequestFields, dst.Fields); err != nil {
		return nil, err
	}

	var out []Employee
	if err := mapstructure.Decode(dst.Employees, &out); err != nil {
		return out, nil
	}

	return out, nil
}

func checkFieldsPresent(want []string, got []field) error {
	fields := make(map[string]struct{}, len(got))
	for _, f := range got {
		fields[f.ID] = struct{}{}
	}

	var missing []string
	for _, f := range want {
		if _, there := fields[f]; !there {
			missing = append(missing, f)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("missing %s fields in the response", strings.Join(missing, ","))
}
