package internal

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type authMiddleware struct {
	Token string
}

// Middleware implements mux.Middleware
func (amw *authMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Token")
		if token == string(amw.Token) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

// NewServer implements new BambooHR limited data exporter
func NewServer(emplReq EmployeeRequest, token string, log zerolog.Logger) http.Handler {
	srv := server{EmployeeRequest: emplReq, Logger: log}

	r := mux.NewRouter()
	r.Path("/employees").Methods(http.MethodGet).HandlerFunc(srv.getEmployees)
	r.Use((&authMiddleware{token}).Middleware)

	return r
}

type server struct {
	EmployeeRequest
	zerolog.Logger
}

func (srv *server) getEmployees(w http.ResponseWriter, r *http.Request) {
	data, err := GetEmployees(r.Context(), srv.EmployeeRequest)
	if err != nil {
		srv.serveError(w, err, "get employees")
		return
	}
	srv.serveJSON(w, data)
}

func (srv *server) serveError(w http.ResponseWriter, err error, msg string) {
	srv.Err(err).Msg(msg)
	w.WriteHeader(http.StatusInternalServerError)
}

func (srv *server) serveJSON(w http.ResponseWriter, src interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		srv.Err(err).Msg("json marshal")
	}
}
