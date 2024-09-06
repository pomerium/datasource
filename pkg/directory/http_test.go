package directory

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	t.Cleanup(clearTimeout)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expect := struct {
			groups []Group
			users  []User
			err    error
		}{}
		h := NewHandler(ProviderFunc(func(_ context.Context) ([]Group, []User, error) {
			return expect.groups, expect.users, expect.err
		}))
		srv := httptest.NewServer(h)
		t.Cleanup(srv.Close)

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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if !assert.NoError(t, err) {
			return
		}

		res, err := http.DefaultClient.Do(req)
		if !assert.NoError(t, err) {
			return
		}
		defer res.Body.Close()

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, `"d3b7677a8420759f"`, res.Header.Get("ETag"))
		groups, users, err := decodeBundle(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, expect.groups, groups)
		assert.Equal(t, expect.users, users)

		req.Header.Set("If-None-Match", `"d3b7677a8420759f"`)
		res, err = http.DefaultClient.Do(req)
		if !assert.NoError(t, err) {
			return
		}
		defer res.Body.Close()

		assert.Equal(t, 304, res.StatusCode)
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()

		expect := struct {
			groups []Group
			users  []User
			err    error
		}{}
		h := NewHandler(ProviderFunc(func(_ context.Context) ([]Group, []User, error) {
			return expect.groups, expect.users, expect.err
		}))
		srv := httptest.NewServer(h)
		t.Cleanup(srv.Close)

		expect.groups = nil
		expect.users = nil
		expect.err = errors.New("ERROR")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if !assert.NoError(t, err) {
			return
		}

		res, err := http.DefaultClient.Do(req)
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
