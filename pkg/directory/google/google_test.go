package google

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/directory"
)

//nolint:gosec
var privateKey = `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

type M = map[string]interface{}

func newMockAPI(t *testing.T, _ *httptest.Server) http.Handler {
	t.Helper()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/token", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(M{
			"access_token":  "ACCESSTOKEN",
			"token_type":    "Bearer",
			"refresh_token": "REFRESHTOKEN",
		})
	})
	r.Route("/admin/directory/v1", func(r chi.Router) {
		r.Route("/groups", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Query().Get("userKey") {
				case "user1":
					_ = json.NewEncoder(w).Encode(M{
						"kind": "admin#directory#groups",
						"groups": []M{
							{"id": "group1"},
							{"id": "group2"},
						},
					})
				default:
					_ = json.NewEncoder(w).Encode(M{
						"kind": "admin#directory#groups",
						"groups": []M{
							{"id": "group1", "directMembersCount": "2"},
							{"id": "group2"},
						},
					})
				}
			})
			r.Get("/{groupKey}/members", func(w http.ResponseWriter, r *http.Request) {
				switch chi.URLParam(r, "groupKey") {
				case "group1":
					_ = json.NewEncoder(w).Encode(M{
						"members": []M{
							{
								"kind":  "admin#directory#member",
								"id":    "inside-user1",
								"email": "user1@inside.test",
								"type":  "USER",
							},
							{
								"kind":  "admin#directory#member",
								"id":    "outside-user1",
								"email": "user1@outside.test",
								"type":  "USER",
							},
						},
					})
				}
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(w).Encode(M{
					"kind": "admin#directory#users",
					"users": []M{
						{
							"kind":         "admin#directory#user",
							"id":           "inside-user1",
							"primaryEmail": "user1@inside.test",
						},
					},
				})
			})
			r.Get("/{user_id}", func(w http.ResponseWriter, r *http.Request) {
				switch chi.URLParam(r, "user_id") {
				case "inside-user1":
					_ = json.NewEncoder(w).Encode(M{
						"kind": "admin#directory#user",
						"id":   "inside-user1",
						"name": M{
							"fullName": "User 1",
						},
						"primaryEmail": "user1@inside.test",
					})
				case "outside-user1":
					http.Error(w, "forbidden", http.StatusForbidden)
				default:
					http.Error(w, "not found", http.StatusNotFound)
				}
			})
		})
	})

	return r
}

func TestProvider_GetDirectory(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	var mockAPI http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockAPI.ServeHTTP(w, r)
	}))
	defer srv.Close()
	mockAPI = newMockAPI(t, srv)

	p := New(
		WithImpersonateUser("IMPERSONATE_USER"),
		WithJSONKey(encodeJSON(map[string]any{
			"type":        "service_account",
			"private_key": privateKey,
			"token_uri":   srv.URL + "/token",
		})),
		WithURL(srv.URL),
	)

	dgs, dus, err := p.GetDirectory(ctx)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, []directory.Group{
		{ID: "group1"},
	}, dgs)
	assert.Equal(t, []directory.User{
		{ID: "inside-user1", Email: "user1@inside.test", GroupIDs: []string{"group1"}},
		{ID: "outside-user1", Email: "user1@outside.test", GroupIDs: []string{"group1"}},
	}, dus)
}

func encodeJSON(data any) []byte {
	bs, _ := json.Marshal(data)
	return bs
}
