package bamboohr_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/bamboohr"
	"github.com/pomerium/datasource/internal/server"
)

func TestAPI(t *testing.T) {
	ctx := context.Background()
	r := mux.NewRouter()
	r.Path("/api/gateway.php/{company}/v1/reports/custom").
		Methods(http.MethodPost).
		HandlerFunc(serveJSON("employees", "company", http.StatusOK))
	srv := httptest.NewServer(r)

	base, err := url.Parse(srv.URL)
	require.NoError(t, err, srv.URL)

	req := bamboohr.EmployeeRequest{
		Auth: bamboohr.Auth{
			BaseURL:   base.ResolveReference(&url.URL{Path: "/api/gateway.php/"}),
			Subdomain: "test",
		},
	}
	client := server.NewDebugClient(http.DefaultClient, zerolog.New(os.Stdout))
	_, err = bamboohr.GetAllEmployees(ctx, client, req)
	require.NoError(t, err, "get employees")
}

func serveJSON(prefix, key string, statusCode int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		value := mux.Vars(r)[key]
		if value == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("expected %s in vars, got none", key)))
			return
		}
		p := path.Join("testdata", fmt.Sprintf("%s-%s-%s.json", prefix, key, value))
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
