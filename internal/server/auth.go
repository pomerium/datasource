package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// TokenMiddleware returns middleware that verifies
func TokenMiddleware(token string) mux.MiddlewareFunc {
	mw := authMiddleware{token}
	return mw.middleware
}

type authMiddleware struct {
	Token string
}

// Middleware implements mux.Middleware
func (amw *authMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Token")
		if token == string(amw.Token) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
