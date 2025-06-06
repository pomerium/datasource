package onelogin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/directory"
)

type M = map[string]interface{}

func newMockAPI(srv *httptest.Server, userIDToGroupName map[int]string) http.Handler {
	lookup := map[string]struct{}{}
	for _, group := range userIDToGroupName {
		lookup[group] = struct{}{}
	}
	var allGroups []string
	for groupName := range lookup {
		allGroups = append(allGroups, groupName)
	}
	sort.Strings(allGroups)

	var allUserIDs []int
	for userID := range userIDToGroupName {
		allUserIDs = append(allUserIDs, userID)
	}
	sort.Ints(allUserIDs)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/auth/oauth2/v2/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "client_id:CLIENTID, client_secret:CLIENTSECRET" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var request struct {
			GrantType string `json:"grant_type"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		if request.GrantType != "client_credentials" {
			http.Error(w, "invalid grant_type", http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(M{
			"access_token":  "ACCESSTOKEN",
			"created_at":    time.Now().Format(time.RFC3339),
			"expires_in":    360000,
			"refresh_token": "REFRESHTOKEN",
			"token_type":    "bearer",
		})
	})
	r.Route("/api/1", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "bearer:ACCESSTOKEN" {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/groups", func(w http.ResponseWriter, r *http.Request) {
			var result struct {
				Pagination struct {
					NextLink string `json:"next_link"`
				} `json:"pagination"`
				Data []M `json:"data"`
			}

			found := r.URL.Query().Get("after") == ""
			for i := range allGroups {
				if found {
					result.Data = append(result.Data, M{
						"id":   i,
						"name": allGroups[i],
					})
					break
				}
				found = r.URL.Query().Get("after") == fmt.Sprint(i)
			}

			if len(result.Data) > 0 {
				nextURL := mustParseURL(srv.URL).ResolveReference(r.URL)
				q := nextURL.Query()
				q.Set("after", fmt.Sprint(result.Data[0]["id"]))
				nextURL.RawQuery = q.Encode()
				result.Pagination.NextLink = nextURL.String()
			}

			_ = json.NewEncoder(w).Encode(result)
		})
		r.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
			userIDToGroupID := map[int]int{}
			for userID, groupName := range userIDToGroupName {
				for id, n := range allGroups {
					if groupName == n {
						userIDToGroupID[userID] = id
					}
				}
			}

			userID, _ := strconv.Atoi(chi.URLParam(r, "user_id"))

			_ = json.NewEncoder(w).Encode(M{
				"data": []M{{
					"id":        userID,
					"email":     userIDToGroupName[userID] + "@example.com",
					"group_id":  userIDToGroupID[userID],
					"firstname": "User",
					"lastname":  fmt.Sprint(userID),
				}},
			})
		})
		r.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
			userIDToGroupID := map[int]int{}
			for userID, groupName := range userIDToGroupName {
				for id, n := range allGroups {
					if groupName == n {
						userIDToGroupID[userID] = id
					}
				}
			}

			var result []M
			for _, userID := range allUserIDs {
				result = append(result, M{
					"id":        userID,
					"email":     userIDToGroupName[userID] + "@example.com",
					"group_id":  userIDToGroupID[userID],
					"firstname": "User",
					"lastname":  fmt.Sprint(userID),
				})
			}
			_ = json.NewEncoder(w).Encode(M{
				"data": result,
			})
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
	mockAPI = newMockAPI(srv, map[int]string{
		111: "admin",
		222: "test",
		333: "user",
	})

	p := New(
		WithClientID("CLIENTID"),
		WithClientSecret("CLIENTSECRET"),
		WithURL(mustParseURL(srv.URL)),
	)
	groups, users, err := p.GetDirectory(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "0", Name: "admin"},
		{ID: "1", Name: "test"},
		{ID: "2", Name: "user"},
	}, groups)
	assert.Equal(t, []directory.User{
		{ID: "111", DisplayName: "User 111", Email: "admin@example.com", GroupIDs: []string{"0"}},
		{ID: "222", DisplayName: "User 222", Email: "test@example.com", GroupIDs: []string{"1"}},
		{ID: "333", DisplayName: "User 333", Email: "user@example.com", GroupIDs: []string{"2"}},
	}, users)
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
