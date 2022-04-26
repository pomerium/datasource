package bamboohr

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// NewServer implements new BambooHR limited data exporter
func NewServer(emplReq EmployeeRequest, client *http.Client, log zerolog.Logger) *mux.Router {
	srv := apiServer{emplReq, client, log}

	r := mux.NewRouter()
	r.Path("/employees/all").Methods(http.MethodGet).HandlerFunc(srv.getAllEmployees)
	r.Path("/employees/available").Methods(http.MethodGet).HandlerFunc(srv.getAvailableEmployees)

	return r
}

type apiServer struct {
	EmployeeRequest
	*http.Client
	zerolog.Logger
}

func (srv *apiServer) getAllEmployees(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	employees, err := GetAllEmployees(ctx, srv.Client, srv.EmployeeRequest)
	if err != nil {
		srv.serveError(w, err, "get employees")
		return
	}

	srv.serveJSON(w, employees)
}

func (srv *apiServer) getAvailableEmployees(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	employees, err := GetAvailableEmployees(ctx, srv.Client, srv.EmployeeRequest)
	if err != nil {
		srv.serveError(w, err, "get employees")
		return
	}

	srv.serveJSON(w, employees)
}

func (srv *apiServer) serveError(w http.ResponseWriter, err error, msg string) {
	srv.Err(err).Msg(msg)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(msg))
	_, _ = w.Write([]byte("\n"))
	_, _ = w.Write([]byte(err.Error()))
}

func (srv *apiServer) serveJSON(w http.ResponseWriter, src interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		srv.Err(err).Msg("json marshal")
	}
}
