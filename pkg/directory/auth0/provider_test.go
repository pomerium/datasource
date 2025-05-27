package auth0_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/auth0"
)

type (
	A = []any
	M = map[string]any
)

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestProvider_GetDirectory(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			assert.Equal(t, "DOMAIN", r.Host)
			switch r.URL.Path {
			case "/api/v2/users":
				page := r.URL.Query().Get("page")
				switch page {
				case "0":
					_ = json.NewEncoder(rec).Encode(M{
						"start": 0,
						"limit": 2,
						"total": 4,
						"users": A{
							M{"user_id": "user1", "name": "User 1", "email": "user1@example.com"},
							M{"user_id": "user2", "name": "User 2", "email": "user2@example.com"},
						},
					})
				case "1":
					_ = json.NewEncoder(rec).Encode(M{
						"start": 2,
						"limit": 2,
						"total": 4,
						"users": A{
							M{"user_id": "user3", "name": "User 3", "email": "user3@example.com"},
							M{"user_id": "user4", "name": "User 4", "email": "user4@example.com"},
						},
					})
				default:
					t.Fatal("unexpected page for user list", page)
				}
			case "/api/v2/roles":
				page := r.URL.Query().Get("page")
				switch page {
				case "0":
					_ = json.NewEncoder(rec).Encode(M{
						"start": 0,
						"limit": 2,
						"total": 4,
						"roles": A{
							M{"id": "team2", "name": "Team 2"},
							M{"id": "team1", "name": "Team 1"},
						},
					})
				case "1":
					_ = json.NewEncoder(rec).Encode(M{
						"start": 2,
						"limit": 2,
						"total": 4,
						"roles": A{
							M{"id": "team4", "name": "Team 4"},
							M{"id": "team3", "name": "Team 3"},
						},
					})
				default:
					t.Fatal("unexpected page for user list", page)
				}
			case "/api/v2/roles/team1/users":
				_ = json.NewEncoder(rec).Encode(M{
					"start": 0,
					"limit": 2,
					"total": 2,
					"users": A{
						M{"user_id": "user1"},
						M{"user_id": "user2"},
					},
				})
			case "/api/v2/roles/team2/users":
				_ = json.NewEncoder(rec).Encode(M{
					"start": 0,
					"limit": 1,
					"total": 1,
					"users": A{
						M{"user_id": "user1"},
					},
				})
			case "/api/v2/roles/team3/users":
				_ = json.NewEncoder(rec).Encode(M{
					"start": 0,
					"limit": 3,
					"total": 3,
					"users": A{
						M{"user_id": "user3"},
						M{"user_id": "user2"},
						M{"user_id": "user1"},
					},
				})
			case "/api/v2/roles/team4/users":
				_ = json.NewEncoder(rec).Encode(M{
					"start": 0,
					"limit": 1,
					"total": 1,
					"users": A{
						M{"user_id": "user4"},
					},
				})
			default:
				rec.WriteHeader(http.StatusNotFound)
			}

			return rec.Result(), nil
		}),
	}

	p := auth0.New(
		auth0.WithClientID("CLIENT_ID"),
		auth0.WithClientSecret("CLIENT_SECRET"),
		auth0.WithDomain("DOMAIN"),
		auth0.WithHTTPClient(httpClient),
		auth0.WithInsecure(true),
	)
	bundle, err := p.GetDirectory(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "team1", Name: "Team 1"},
		{ID: "team2", Name: "Team 2"},
		{ID: "team3", Name: "Team 3"},
		{ID: "team4", Name: "Team 4"},
	}, bundle.Groups())
	assert.Equal(t, []directory.User{
		{ID: "user1", GroupIDs: []string{"team1", "team2", "team3"}, DisplayName: "User 1", Email: "user1@example.com"},
		{ID: "user2", GroupIDs: []string{"team1", "team3"}, DisplayName: "User 2", Email: "user2@example.com"},
		{ID: "user3", GroupIDs: []string{"team3"}, DisplayName: "User 3", Email: "user3@example.com"},
		{ID: "user4", GroupIDs: []string{"team4"}, DisplayName: "User 4", Email: "user4@example.com"},
	}, bundle.Users())
}
