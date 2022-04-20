package zenefits_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/zenefits"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	ctx := context.Background()
	r := mux.NewRouter()
	r.Path("/core/people").
		Methods(http.MethodGet).
		HandlerFunc(serveJSON("testdata/people.json", http.StatusOK))
	r.Use(server.AuthorizationBearerMiddleware("test"))
	srv := httptest.NewServer(r)

	base, err := url.Parse(srv.URL)
	require.NoError(t, err, srv.URL)

	client := server.NewBearerTokenClient(http.DefaultClient, "test")
	req := zenefits.PeopleRequest{
		Auth: zenefits.Auth{
			BaseURL: base.ResolveReference(&url.URL{Path: "/core/"}),
		},
	}
	resp, err := zenefits.GetEmployees(ctx, client, req)
	require.NoError(t, err, "get employees")

	var dst []map[string]interface{}
	require.NoError(t, mapstructure.Decode(resp, &dst))
	_, err = json.Marshal(dst)
	require.NoError(t, err)
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
