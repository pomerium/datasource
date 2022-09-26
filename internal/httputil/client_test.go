package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		var clearTimeout context.CancelFunc
		ctx, clearTimeout = context.WithDeadline(ctx, deadline)
		t.Cleanup(clearTimeout)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/200":
			http.Error(w, "OK", http.StatusOK)
		case "/500":
			http.Error(w, "NOT OK", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	for _, tc := range []struct {
		name             string
		path             string
		expectBody       string
		expectStatusCode int
	}{
		{"ok", "/200", "OK\n", 200},
		{"error", "/500", "NOT OK\n", 500},
		{"not-found", "/400", "404 page not found\n", 404},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := zerolog.New(&buf)
			client := NewLoggingClient(logger, nil, func(event *zerolog.Event) *zerolog.Event {
				return event.Str("test", "TEST")
			})
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+tc.path, nil)
			require.NoError(t, err)
			res, err := client.Do(req)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectStatusCode, res.StatusCode)
				bs, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectBody, string(bs))
				err = res.Body.Close()
				assert.NoError(t, err)
			}

			actualLog := map[string]any{}
			assert.NoError(t, json.Unmarshal(buf.Bytes(), &actualLog))
			expectLog := map[string]any{
				"authority":     actualLog["authority"],
				"duration":      actualLog["duration"],
				"level":         "debug",
				"message":       "http-request",
				"method":        "GET",
				"path":          tc.path,
				"response-code": float64(tc.expectStatusCode),
				"test":          "TEST",
			}
			if tc.expectStatusCode != 200 {
				expectLog["response-body"] = tc.expectBody
			}
			assert.Equal(t, expectLog, actualLog)
		})
	}
}
