package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/pomerium/datasource/pkg/directory"
)

type M = map[string]interface{}

func newMockAPI(t *testing.T, _ *httptest.Server) http.Handler {
	t.Helper()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !assert.Equal(t, "Basic YWJjOnh5eg==", r.Header.Get("Authorization")) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Post("/graphql", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		q, err := parser.ParseQuery(&ast.Source{
			Input: body.Query,
		})
		if err != nil {
			panic(err)
		}

		result := qlResult{
			Data: &qlData{
				Organization: &qlOrganization{},
			},
		}
		handleMembersWithRole := func(orgSlug string, field *ast.Field) {
			membersWithRole := &qlMembersWithRoleConnection{}

			var cursor string
			for _, arg := range field.Arguments {
				if arg.Name == "after" {
					cursor = arg.Value.Raw
				}
			}

			switch cursor {
			case `null`:
				switch orgSlug {
				case "org1":
					membersWithRole.PageInfo = qlPageInfo{EndCursor: "TOKEN1", HasNextPage: true}
					membersWithRole.Nodes = []qlUser{
						{ID: "user1", Login: "user1", Name: "User 1", Email: "user1@example.com"},
						{ID: "user2", Login: "user2", Name: "User 2", Email: "user2@example.com"},
					}
				case "org2":
					membersWithRole.PageInfo = qlPageInfo{HasNextPage: false}
					membersWithRole.Nodes = []qlUser{
						{ID: "user4", Login: "user4", Name: "User 4", Email: "user4@example.com"},
					}
				default:
					t.Errorf("unexpected org slug: %s", orgSlug)
				}
			case `TOKEN1`:
				membersWithRole.PageInfo = qlPageInfo{HasNextPage: false}
				membersWithRole.Nodes = []qlUser{
					{ID: "user3", Login: "user3", Name: "User 3", Email: "user3@example.com"},
				}
			default:
				t.Errorf("unexpected cursor: %s", cursor)
			}

			result.Data.Organization.MembersWithRole = membersWithRole
		}
		handleTeamMembers := func(_, teamSlug string, _ *ast.Field) {
			result.Data.Organization.Team.Members = &qlTeamMemberConnection{
				PageInfo: qlPageInfo{HasNextPage: false},
			}

			switch teamSlug {
			case "team3":
				result.Data.Organization.Team.Members.Edges = []qlTeamMemberEdge{
					{Node: qlUser{ID: "user3"}},
				}
			}
		}
		handleTeam := func(orgSlug string, field *ast.Field) {
			result.Data.Organization.Team = &qlTeam{}

			var teamSlug string
			for _, arg := range field.Arguments {
				if arg.Name == "slug" {
					teamSlug = arg.Value.Raw
				}
			}

			for _, selection := range field.SelectionSet {
				subField, ok := selection.(*ast.Field)
				if !ok {
					continue
				}

				switch subField.Name {
				case "members":
					handleTeamMembers(orgSlug, teamSlug, subField)
				}
			}
		}
		renderNodeField := func(field *ast.Field, path []string, value string) string {
		outer:
			for _, segment := range path {
				for _, selection := range field.SelectionSet {
					subField, ok := selection.(*ast.Field)
					if !ok {
						continue
					}

					if subField.Name != segment {
						continue
					}

					field = subField
					continue outer
				}
				return ""
			}
			return value
		}
		handleTeams := func(orgSlug string, field *ast.Field) {
			teams := &qlTeamConnection{}

			var cursor string
			var userLogin string
			for _, arg := range field.Arguments {
				if arg.Name == "after" {
					cursor = arg.Value.Raw
				}
				if arg.Name == "userLogins" {
					userLogin = arg.Value.Children[0].Value.Raw
				}
			}

			switch cursor {
			case `null`:
				switch orgSlug {
				case "org1":
					teams.PageInfo = qlPageInfo{HasNextPage: true, EndCursor: "TOKEN1"}
					teams.Edges = []qlTeamEdge{
						{Node: qlTeam{
							ID:   renderNodeField(field, []string{"edges", "node", "id"}, "MDQ6VGVhbTE="),
							Slug: renderNodeField(field, []string{"edges", "node", "slug"}, "team1"),
							Name: renderNodeField(field, []string{"edges", "node", "name"}, "Team 1"),
							Members: &qlTeamMemberConnection{
								PageInfo: qlPageInfo{HasNextPage: false},
								Edges: []qlTeamMemberEdge{
									{Node: qlUser{ID: "user1"}},
									{Node: qlUser{ID: "user2"}},
								},
							},
						}},
					}
				case "org2":
					teams.PageInfo = qlPageInfo{HasNextPage: false}
					teams.Edges = []qlTeamEdge{
						{Node: qlTeam{
							ID:   renderNodeField(field, []string{"edges", "node", "id"}, "MDQ6VGVhbTM="),
							Slug: renderNodeField(field, []string{"edges", "node", "slug"}, "team3"),
							Name: renderNodeField(field, []string{"edges", "node", "name"}, "Team 3"),
							Members: &qlTeamMemberConnection{
								PageInfo: qlPageInfo{HasNextPage: true, EndCursor: "TOKEN1"},
								Edges: []qlTeamMemberEdge{
									{Node: qlUser{ID: "user1"}},
									{Node: qlUser{ID: "user2"}},
								},
							},
						}},
					}
					if userLogin == "" || userLogin == "user4" {
						teams.Edges = append(teams.Edges, qlTeamEdge{
							Node: qlTeam{
								ID:   renderNodeField(field, []string{"edges", "node", "id"}, "MDQ6VGVhbTQ="),
								Slug: renderNodeField(field, []string{"edges", "node", "slug"}, "team4"),
								Name: renderNodeField(field, []string{"edges", "node", "name"}, "Team 4"),
								Members: &qlTeamMemberConnection{
									PageInfo: qlPageInfo{HasNextPage: false},
									Edges: []qlTeamMemberEdge{
										{Node: qlUser{ID: "user4"}},
									},
								},
							},
						})
					}
				default:
					t.Errorf("unexpected org slug: %s", orgSlug)
				}
			case "TOKEN1":
				teams.PageInfo = qlPageInfo{HasNextPage: false}
				teams.Edges = []qlTeamEdge{
					{Node: qlTeam{
						ID:   renderNodeField(field, []string{"edges", "node", "id"}, "MDQ6VGVhbTI="),
						Slug: renderNodeField(field, []string{"edges", "node", "slug"}, "team2"),
						Name: renderNodeField(field, []string{"edges", "node", "name"}, "Team 2"),
						Members: &qlTeamMemberConnection{
							PageInfo: qlPageInfo{HasNextPage: false},
							Edges: []qlTeamMemberEdge{
								{Node: qlUser{ID: "user1"}},
							},
						},
					}},
				}
			default:
				t.Errorf("unexpected cursor: %s", cursor)
			}

			result.Data.Organization.Teams = teams
		}
		handleOrganization := func(field *ast.Field) {
			var orgSlug string
			for _, arg := range field.Arguments {
				if arg.Name == "login" {
					orgSlug = arg.Value.Raw
				}
			}
			for _, orgSelection := range field.SelectionSet {
				orgField, ok := orgSelection.(*ast.Field)
				if !ok {
					continue
				}

				switch orgField.Name {
				case "teams":
					handleTeams(orgSlug, orgField)
				case "team":
					handleTeam(orgSlug, orgField)
				case "membersWithRole":
					handleMembersWithRole(orgSlug, orgField)
				}
			}
		}

		for _, operation := range q.Operations {
			for _, selection := range operation.SelectionSet {
				field, ok := selection.(*ast.Field)
				if !ok {
					continue
				}

				if field.Name != "organization" {
					continue
				}

				handleOrganization(field)
			}
		}

		_ = json.NewEncoder(w).Encode(result)
	})
	r.Get("/user/orgs", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]M{
			{"login": "org1"},
			{"login": "org2"},
		})
	})
	r.Get("/orgs/{org_id}/teams", func(w http.ResponseWriter, r *http.Request) {
		teams := map[string][]M{
			"org1": {
				{"slug": "team1", "id": 1},
				{"slug": "team2", "id": 2},
			},
			"org2": {
				{"slug": "team3", "id": 3},
				{"slug": "team4", "id": 4},
			},
		}
		orgID := chi.URLParam(r, "org_id")
		_ = json.NewEncoder(w).Encode(teams[orgID])
	})
	r.Get("/orgs/{org_id}/teams/{team_id}/members", func(w http.ResponseWriter, r *http.Request) {
		members := map[string]map[string][]M{
			"org1": {
				"team1": {
					{"login": "user1"},
					{"login": "user2"},
				},
				"team2": {
					{"login": "user1"},
				},
			},
			"org2": {
				"team3": {
					{"login": "user1"},
					{"login": "user2"},
					{"login": "user3"},
				},
				"team4": {
					{"login": "user4"},
				},
			},
		}
		orgID := chi.URLParam(r, "org_id")
		teamID := chi.URLParam(r, "team_id")
		_ = json.NewEncoder(w).Encode(members[orgID][teamID])
	})
	r.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		users := map[string]apiUserObject{
			"user1": {Login: "user1", Name: "User 1", Email: "user1@example.com"},
			"user2": {Login: "user2", Name: "User 2", Email: "user2@example.com"},
			"user3": {Login: "user3", Name: "User 3", Email: "user3@example.com"},
			"user4": {Login: "user4", Name: "User 4", Email: "user4@example.com"},
		}
		userID := chi.URLParam(r, "user_id")
		_ = json.NewEncoder(w).Encode(users[userID])
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
		WithPersonalAccessToken("xyz"),
		WithURL(mustParseURL(srv.URL)),
		WithUsername("abc"),
	)
	bundle, err := p.GetDirectory(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "team1", Name: "team1"},
		{ID: "team2", Name: "team2"},
		{ID: "team3", Name: "team3"},
		{ID: "team4", Name: "team4"},
	}, bundle.Groups())
	assert.Equal(t, []directory.User{
		{ID: "user1", GroupIDs: []string{"team1", "team2", "team3"}, DisplayName: "User 1", Email: "user1@example.com"},
		{ID: "user2", GroupIDs: []string{"team1", "team3"}, DisplayName: "User 2", Email: "user2@example.com"},
		{ID: "user3", GroupIDs: []string{"team3"}, DisplayName: "User 3", Email: "user3@example.com"},
		{ID: "user4", GroupIDs: []string{"team4"}, DisplayName: "User 4", Email: "user4@example.com"},
	}, bundle.Users())
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
