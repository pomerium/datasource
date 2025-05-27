package azure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/directory"
)

type M = map[string]any

func newMockAPI(t *testing.T, _ *httptest.Server) http.Handler {
	t.Helper()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	tokenCount := 0
	r.Post("/DIRECTORY_ID/oauth2/v2.0/token", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "CLIENT_ID", r.FormValue("client_id"))
		assert.Equal(t, "CLIENT_SECRET", r.FormValue("client_secret"))
		assert.Equal(t, defaultLoginScope, r.FormValue("scope"))
		assert.Equal(t, defaultLoginGrantType, r.FormValue("grant_type"))
		tokenCount++

		_ = json.NewEncoder(w).Encode(M{
			"access_token":  fmt.Sprintf("ACCESSTOKEN%d", tokenCount),
			"token_type":    "Bearer",
			"refresh_token": "REFRESHTOKEN",
		})
	})
	r.Route("/v1.0", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Header.Get("Authorization") {
				case "Bearer ACCESSTOKEN1":
					http.Error(w, "expired", http.StatusUnauthorized)
				case "Bearer ACCESSTOKEN2":
					next.ServeHTTP(w, r)
				default:
					http.Error(w, "forbidden", http.StatusForbidden)
				}
			})
		})
		r.Get("/groups/delta", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(M{
				"value": []M{
					{
						"id":          "admin",
						"displayName": "Admin Group",
						"members@delta": []M{
							{"@odata.type": "#microsoft.graph.user", "id": "user-1"},
							{"@odata.type": "#microsoft.graph.servicePrincipal", "id": "service-principal-1"},
						},
					},
					{
						"id":          "test",
						"displayName": "Test Group",
						"members@delta": []M{
							{"@odata.type": "#microsoft.graph.user", "id": "user-2"},
							{"@odata.type": "#microsoft.graph.user", "id": "user-3"},
							{"@odata.type": "#microsoft.graph.user", "id": "user-4"},
						},
					},
				},
			})
		})
		r.Get("/servicePrincipals/delta", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(M{
				"value": []M{
					{
						"id":          "service-principal-1",
						"displayName": "Service Principal 1",
					},
					{
						"id":          "service-principal-2",
						"displayName": "Service Principal 2",
					},
				},
			})
		})
		r.Get("/users/delta", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(M{
				"value": []M{
					{"id": "user-1", "displayName": "User 1", "mail": "user1@example.com"},
					{"id": "user-2", "displayName": "User 2", "mail": "user2@example.com"},
					{"id": "user-3", "displayName": "User 3", "userPrincipalName": "user3_example.com#EXT#@user3example.onmicrosoft.com"},
					{"id": "user-4", "displayName": "User 4", "userPrincipalName": "user4@example.com"},
				},
			})
		})
		r.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
			switch chi.URLParam(r, "user_id") {
			case "user-1":
				_ = json.NewEncoder(w).Encode(M{"id": "user-1", "displayName": "User 1", "mail": "user1@example.com"})
			default:
				http.Error(w, "not found", http.StatusNotFound)
			}
		})
		r.Get("/users/{user_id}/transitiveMemberOf", func(w http.ResponseWriter, r *http.Request) {
			switch chi.URLParam(r, "user_id") {
			case "user-1":
				switch r.URL.Query().Get("page") {
				case "":
					_ = json.NewEncoder(w).Encode(M{
						"value": []M{
							{"id": "admin"},
						},
						"@odata.nextLink": getPageURL(r, 1),
					})
				case "1":
					_ = json.NewEncoder(w).Encode(M{
						"value": []M{
							{"id": "group1"},
						},
						"@odata.nextLink": getPageURL(r, 2),
					})
				case "2":
					_ = json.NewEncoder(w).Encode(M{
						"value": []M{
							{"id": "group2"},
						},
					})
				}
			default:
				http.Error(w, "not found", http.StatusNotFound)
			}
		})
	})

	return r
}

func TestProvider_GetDirectory(t *testing.T) {
	t.Parallel()

	var mockAPI http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockAPI.ServeHTTP(w, r)
	}))
	defer srv.Close()
	mockAPI = newMockAPI(t, srv)

	p := New(
		WithClientID("CLIENT_ID"),
		WithClientSecret("CLIENT_SECRET"),
		WithDirectoryID("DIRECTORY_ID"),
		WithGraphURL(mustParseURL(srv.URL)),
		WithLoginURL(mustParseURL(srv.URL)),
	)
	bundle, err := p.GetDirectory(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "admin", Name: "Admin Group"},
		{ID: "test", Name: "Test Group"},
	}, bundle.Groups())
	assert.Equal(t, []directory.User{
		{
			ID:          "service-principal-1",
			GroupIDs:    []string{"admin"},
			DisplayName: "Service Principal 1",
		},
		{
			ID:          "service-principal-2",
			GroupIDs:    []string{},
			DisplayName: "Service Principal 2",
		},
		{
			ID:          "user-1",
			GroupIDs:    []string{"admin"},
			DisplayName: "User 1",
			Email:       "user1@example.com",
		},
		{
			ID:          "user-2",
			GroupIDs:    []string{"test"},
			DisplayName: "User 2",
			Email:       "user2@example.com",
		},
		{
			ID:          "user-3",
			GroupIDs:    []string{"test"},
			DisplayName: "User 3",
			Email:       "user3@example.com",
		},
		{
			ID:          "user-4",
			GroupIDs:    []string{"test"},
			DisplayName: "User 4",
			Email:       "user4@example.com",
		},
	}, bundle.Users())
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func getPageURL(r *http.Request, page int) string {
	u := *r.URL
	if r.TLS == nil {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = r.Host
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}
