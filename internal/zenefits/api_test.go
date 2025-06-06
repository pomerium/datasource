package zenefits_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/zenefits"
)

func TestAPI(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := mux.NewRouter()
	r.Path("/core/people").
		Methods(http.MethodGet).
		HandlerFunc(serveJSON("testdata/people.json", http.StatusOK))
	r.Path("/time_off/vacation_requests").
		Methods(http.MethodGet).
		HandlerFunc(serveJSON("testdata/vacations.json", http.StatusOK))
	srv := httptest.NewServer(r)

	base, err := url.Parse(srv.URL)
	require.NoError(t, err, srv.URL)

	client := server.NewDebugClient(http.DefaultClient, zerolog.New(os.Stdout))
	auth := zenefits.Auth{
		BaseURL: base,
	}

	t.Run("people", func(t *testing.T) {
		t.Parallel()

		resp, err := zenefits.GetEmployees(ctx, client, zenefits.PeopleRequest{Auth: auth})
		require.NoError(t, err, "get employees")

		var dst []map[string]interface{}
		require.NoError(t, mapstructure.Decode(resp, &dst))
		_, err = json.Marshal(dst)
		require.NoError(t, err)
	})

	t.Run("vacations", func(t *testing.T) {
		t.Parallel()

		resp, err := zenefits.GetVacations(ctx, client, zenefits.VacationRequest{
			Auth:  auth,
			Start: time.Date(2002, 0o5, 0o6, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2002, 0o5, 0o6, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err, "get vacations")
		_, there := resp["26455996"]
		assert.True(t, there)
	})
}

func serveJSON(p string, statusCode int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		data, err := os.ReadFile(p)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write(data)
	}
}
