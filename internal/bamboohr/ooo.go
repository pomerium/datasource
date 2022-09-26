package bamboohr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/pomerium/datasource/internal/util"
)

type WhoIsOutRequest struct {
	Auth
	// Location is required as BambooHR only returns date, and it must be converted into timestamp
	// Timezone is an installation specific setting that does not seem to be available via API
	// but may only be accessed from the BambooHR Admin UI
	Location   *time.Location
	Start, End time.Time
}

const BambooDateLayout = "2006-01-02"

func (req WhoIsOutRequest) RequestURL() *url.URL {
	u := req.Auth.RequestURL("v1/time_off/whos_out")

	vals := make(url.Values)
	vals.Add("start", req.Start.Format(BambooDateLayout))
	vals.Add("end", req.End.Format(BambooDateLayout))
	u.RawQuery = vals.Encode()

	return u
}

// GetAvailableEmployees only returns employees that are marked as active
// and are not on vacation or absence leave
func GetAvailableEmployees(ctx context.Context, client *http.Client, param EmployeeRequest) ([]Employee, error) {
	employees, err := GetAllEmployees(ctx, client, param)
	if err != nil {
		return nil, fmt.Errorf("get employees: %w", err)
	}

	ooo, err := WhoIsOut(ctx, client, WhoIsOutRequest{
		Auth:     param.Auth,
		Location: param.Location,
		Start:    time.Now(),
		End:      time.Now().Add(time.Hour * 24),
	})
	if err != nil {
		return nil, fmt.Errorf("who is out: %w", err)
	}

	return filterOOO(employees, ooo), nil
}

// WhoIsOut retrieves list of employees who are currently marked as out
func WhoIsOut(ctx context.Context, client *http.Client, param WhoIsOutRequest) (map[string][]Period, error) {
	u := param.RequestURL()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
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
		return nil, fmt.Errorf("GET %v: unexpected return status: %s", u, resp.Status)
	}

	p, err := parseWhoIsOutResponse(resp.Body, param.Location)
	if err != nil {
		return nil, fmt.Errorf("who is out response: %w", err)
	}
	return p, nil
}

type Period struct {
	Start time.Time
	End   time.Time
}

// returns map of employeeID to period when one is out
func parseWhoIsOutResponse(r io.Reader, location *time.Location) (map[string][]Period, error) {
	var objs []map[string]interface{}
	if err := json.NewDecoder(r).Decode(&objs); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	dst := make([]struct {
		Type       string        `json:"type" mapstructure:"time"`
		EmployeeID json.Number   `json:"employeeId" mapstructure:"employeeId"`
		Start      util.DateTime `json:"start" mapstructure:"start"`
		End        util.DateTime `json:"end" mapstructure:"end"`
	}, 0, len(objs))
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			util.DateTimeDecodeHook(BambooDateLayout, location),
			util.JSONNumberDecodeHook,
		),
		Result: &dst,
	})
	if err != nil {
		return nil, fmt.Errorf("mapstructure decoder: %w", err)
	}

	if err = dec.Decode(objs); err != nil {
		return nil, fmt.Errorf("mapstructure decode: %w", err)
	}

	out := make(map[string][]Period)
	for _, rec := range dst {
		id := string(rec.EmployeeID)
		if id == "" {
			// some other period kind that is not employee specific,
			// and cannot be resolved
			continue
		}
		out[id] = append(out[id], Period{
			Start: rec.Start.Time(),
			End:   rec.End.Time().Add(time.Hour * 24),
		})
	}

	return out, nil
}

func isOut(now time.Time, out []Period) bool {
	for _, p := range out {
		if now.After(p.Start) && now.Before(p.End) {
			return true
		}
	}
	return false
}

func filterOOO(src []Employee, ooo map[string][]Period) []Employee {
	dst := make([]Employee, 0, len(src))
	now := time.Now()
	for _, emp := range src {
		if !isOut(now, ooo[emp.ID.String()]) {
			dst = append(dst, emp)
		}
	}

	return dst
}
