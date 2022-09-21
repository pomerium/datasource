package directory

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	expect := struct {
		groups []Group
		users  []User
		err    error
	}{}
	h := NewHandler(ProviderFunc(func(ctx context.Context) ([]Group, []User, error) {
		return expect.groups, expect.users, expect.err
	}))

	srv := httptest.NewServer(h)
	defer srv.Close()

	t.Run("success", func(t *testing.T) {
		expect.groups = []Group{
			{ID: "group1", Name: "Group 1"},
			{ID: "group2", Name: "Group 2"},
			{ID: "group3", Name: "Group 4"},
		}
		expect.users = []User{
			{ID: "user1", Email: "user1@example.com"},
			{ID: "user2", Email: "user2@example.com"},
			{ID: "user3", Email: "user3@example.com"},
		}
		expect.err = nil

		res, err := http.Get(srv.URL)
		if !assert.NoError(t, err) {
			return
		}
		defer res.Body.Close()

		assert.Equal(t, 200, res.StatusCode)
		groups, users, err := decodeBundle(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, expect.groups, groups)
		assert.Equal(t, expect.users, users)
	})
	t.Run("error", func(t *testing.T) {
		expect.groups = nil
		expect.users = nil
		expect.err = errors.New("ERROR")

		res, err := http.Get(srv.URL)
		if !assert.NoError(t, err) {
			return
		}
		defer res.Body.Close()

		assert.Equal(t, 500, res.StatusCode)
		bs, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "failed to get directory data: ERROR", strings.TrimSpace(string(bs)))
	})
}
