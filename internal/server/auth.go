package server

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"

	"github.com/pomerium/datasource/internal"
)

// AuthorizationBearerMiddleware enforces Authorization: Bearer token
func AuthorizationBearerMiddleware(token string) mux.MiddlewareFunc {
	mw := bearerMiddleware{token}
	return mw.middleware
}

type bearerMiddleware struct {
	Token string
}

var (
	reAuthorizationBearer = regexp.MustCompile(`^Bearer (\S+)$`)
)

func (mw *bearerMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match := reAuthorizationBearer.FindStringSubmatch(r.Header.Get("Authorization"))
		if len(match) != 2 {
			http.Error(w, "Authorization: Bearer token required", http.StatusUnauthorized)
			return
		}

		if match[1] == string(mw.Token) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

type transport struct {
	token string
	rt    http.RoundTripper
}

// RoundTrip executes a single HTTP transaction
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", internal.ProjectName, internal.Version))
	return t.rt.RoundTrip(req)
}

func NewBearerTokenClient(base *http.Client, token string) *http.Client {
	client := *base
	tr := &transport{token, client.Transport}
	if tr.rt == nil {
		tr.rt = http.DefaultTransport
	}
	client.Transport = tr
	return &client
}
