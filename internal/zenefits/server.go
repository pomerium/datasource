package zenefits

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/pomerium/datasource/internal/util"
	"github.com/rs/zerolog"
)

// NewServer implements new Zenefits limited data exporter
func NewServer(req PeopleRequest, client *http.Client, log zerolog.Logger) *mux.Router {
	srv := apiServer{req, client, log}

	r := mux.NewRouter()
	r.Path("/employees").Methods(http.MethodGet).HandlerFunc(srv.serveEmployees)

	return r
}

type apiServer struct {
	PeopleRequest
	*http.Client
	zerolog.Logger
}

func (srv *apiServer) serveEmployees(w http.ResponseWriter, r *http.Request) {
	data, err := srv.getEmployeesJSON(r.Context())
	if err != nil {
		srv.serveError(w, err, "get employees")
		return
	}

	srv.serveJSON(w, data)
}

func (srv *apiServer) getEmployeesJSON(ctx context.Context) ([]map[string]interface{}, error) {
	persons, err := GetEmployees(ctx, srv.Client, srv.PeopleRequest)
	if err != nil {
		return nil, fmt.Errorf("api: %w", err)
	}

	var dst []map[string]interface{}
	if err = mapstructure.Decode(persons, &dst); err != nil {
		return nil, fmt.Errorf("transform response: %w", err)
	}

	if len(srv.PeopleRequest.Fields) > 0 {
		util.Filter(dst, srv.PeopleRequest.Fields)
	}

	return dst, nil
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
