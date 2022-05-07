package zenefits_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/zenefits"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	ctx := context.Background()
	r := mux.NewRouter()
	r.Path("/core/people").
		Methods(http.MethodGet).
		HandlerFunc(serveJSON("testdata/people.json", http.StatusOK))
	r.Path("/time_off/vacation_requests").
		Methods(http.MethodGet).
		HandlerFunc(serveJSON("testdata/vacations.json", http.StatusOK))
	r.Use(server.AuthorizationBearerMiddleware("test"))
	srv := httptest.NewServer(r)

	base, err := url.Parse(srv.URL)
	require.NoError(t, err, srv.URL)

	client := server.NewBearerTokenClient(http.DefaultClient, "test")
	client = server.NewDebugClient(client, zerolog.New(os.Stdout))
	auth := zenefits.Auth{
		BaseURL: base,
	}

	t.Run("people", func(t *testing.T) {
		resp, err := zenefits.GetEmployees(ctx, client, zenefits.PeopleRequest{Auth: auth})
		require.NoError(t, err, "get employees")

		var dst []map[string]interface{}
		require.NoError(t, mapstructure.Decode(resp, &dst))
		_, err = json.Marshal(dst)
		require.NoError(t, err)
	})

	t.Run("vacations", func(t *testing.T) {
		resp, err := zenefits.GetVacations(ctx, client, zenefits.VacationRequest{
			Auth:  auth,
			Start: time.Date(2002, 05, 06, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2002, 05, 06, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err, "get vacations")
		_, there := resp["26455996"]
		assert.True(t, there)
	})
}

func serveJSON(p string, statusCode int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
