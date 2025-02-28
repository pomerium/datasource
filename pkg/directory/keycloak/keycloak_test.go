package keycloak_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/keycloak"
)

type (
	A = []any
	M = map[string]any
)

func newMockAPI(t *testing.T, srv *httptest.Server) http.Handler {
	t.Helper()

	groups := A{
		M{"id": "g1", "name": "group-1"},
		M{"id": "g2", "name": "group-2"},
		M{"id": "g3", "name": "group-3"},
		M{"id": "g4", "name": "group-4"},
		M{"id": "g5", "name": "group-5"},
	}
	users := A{
		M{"id": "u1", "email": "u1@example.com", "emailVerified": true, "username": "user-1"},
		M{"id": "u2", "email": "u2@example.com", "emailVerified": true, "username": "user-2"},
		M{"id": "u3", "email": "u3@example.com", "emailVerified": true, "username": "user-3"},
		M{"id": "u4", "email": "u4@example.com", "username": "user-4"},
	}
	lookup := map[string]A{
		"g1": {M{"id": "u1"}, M{"id": "u2"}},
		"g2": {M{"id": "u3"}},
		"g3": {M{"id": "u1"}, M{"id": "u2"}, M{"id": "u3"}},
	}

	sendList := func(w http.ResponseWriter, r *http.Request, lst A) {
		o, err := strconv.Atoi(r.FormValue("first"))
		require.NoError(t, err)
		sz, err := strconv.Atoi(r.FormValue("max"))
		require.NoError(t, err)
		next := lst[o:]
		next = next[:min(len(next), sz)]
		w.Header().Set("Content-Type", "application/json")
		if next == nil {
			next = A{}
		}
		_ = json.NewEncoder(w).Encode(next)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /realms/REALM/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(M{
			"issuer":         srv.URL + "/realms/REALM",
			"token_endpoint": srv.URL + "/realms/REALM/protocol/openid-connect/token",
		})
	})
	mux.HandleFunc("POST /realms/REALM/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "CLIENT_ID", r.FormValue("client_id"))
		assert.Equal(t, "CLIENT_SECRET", r.FormValue("client_secret"))
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(M{
			"access_token": "ACCESS_TOKEN",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, groups)
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups/g1/members", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, lookup["g1"])
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups/g2/members", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, lookup["g2"])
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups/g3/members", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, lookup["g3"])
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups/g4/members", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, lookup["g4"])
	})
	mux.HandleFunc("GET /admin/realms/REALM/groups/g5/members", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, lookup["g5"])
	})
	mux.HandleFunc("GET /admin/realms/REALM/users", func(w http.ResponseWriter, r *http.Request) {
		sendList(w, r, users)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.Method, r.URL.String())
		mux.ServeHTTP(w, r)
	})
}

func TestKeyCloak(t *testing.T) {
	t.Parallel()

	var mockAPI http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockAPI.ServeHTTP(w, r)
	}))
	defer srv.Close()
	mockAPI = newMockAPI(t, srv)

	k := keycloak.New(
		keycloak.WithBatchSize(3),
		keycloak.WithClientID("CLIENT_ID"),
		keycloak.WithClientSecret("CLIENT_SECRET"),
		keycloak.WithRealm("REALM"),
		keycloak.WithURL(srv.URL),
	)
	dgs, dus, err := k.GetDirectory(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "g1", Name: "group-1"},
		{ID: "g2", Name: "group-2"},
		{ID: "g3", Name: "group-3"},
		{ID: "g4", Name: "group-4"},
		{ID: "g5", Name: "group-5"},
	}, dgs)
	assert.Equal(t, []directory.User{
		{ID: "u1", DisplayName: "user-1", Email: "u1@example.com", GroupIDs: []string{"g1", "g3"}},
		{ID: "u2", DisplayName: "user-2", Email: "u2@example.com", GroupIDs: []string{"g1", "g3"}},
		{ID: "u3", DisplayName: "user-3", Email: "u3@example.com", GroupIDs: []string{"g2", "g3"}},
		{ID: "u4", DisplayName: "user-4"},
	}, dus)
}
