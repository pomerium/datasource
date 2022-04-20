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
	r.Path("/employees").Methods(http.MethodGet).HandlerFunc(srv.getEmployees)

	return r
}

type apiServer struct {
	EmployeeRequest
	*http.Client
	zerolog.Logger
}

func (srv *apiServer) getEmployees(w http.ResponseWriter, r *http.Request) {
	data, err := GetEmployees(r.Context(), srv.Client, srv.EmployeeRequest)
	if err != nil {
		srv.serveError(w, err, "get employees")
		return
	}
	srv.serveJSON(w, data)
}

func (srv *apiServer) serveError(w http.ResponseWriter, err error, msg string) {
	srv.Err(err).Msg(msg)
	w.WriteHeader(http.StatusInternalServerError)
}

func (srv *apiServer) serveJSON(w http.ResponseWriter, src interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		srv.Err(err).Msg("json marshal")
	}
}
