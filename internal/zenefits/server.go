package zenefits

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

// NewServer implements new Zenefits limited data exporter
func NewServer(req PeopleRequest, client *http.Client, options ...Option) *mux.Router {
	srv := &apiServer{pr: req, client: client, log: zerolog.Nop()}

	for _, opt := range options {
		opt(srv)
	}

	r := mux.NewRouter()
	r.Path("/employees").Methods(http.MethodGet).HandlerFunc(srv.serveEmployees)

	return r
}

type apiServer struct {
	pr               PeopleRequest
	removeOnVacation bool
	location         *time.Location
	client           *http.Client
	log              zerolog.Logger
}

// Option to customize
type Option func(*apiServer)

func WithLogger(log zerolog.Logger) Option {
	return func(as *apiServer) {
		as.log = log
	}
}

func WithRemoveOnVacation(location *time.Location) Option {
	return func(as *apiServer) {
		as.location = location
		as.removeOnVacation = true
	}
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
	persons, err := GetEmployees(ctx, srv.client, srv.pr)
	if err != nil {
		return nil, fmt.Errorf("get people: %w", err)
	}

	if srv.removeOnVacation {
		persons, err = srv.filterOOO(ctx, persons)
		if err != nil {
			return nil, fmt.Errorf("get vacation: %w", err)
		}
	}

	var dst []map[string]interface{}
	if err = mapstructure.Decode(persons, &dst); err != nil {
		return nil, fmt.Errorf("transform response: %w", err)
	}

	return dst, nil
}

func (srv *apiServer) filterOOO(ctx context.Context, persons []Person) ([]Person, error) {
	ooo, err := GetVacations(ctx, srv.client, VacationRequest{
		Auth:  srv.pr.Auth,
		Start: time.Now().In(srv.location),
		End:   time.Now().In(srv.location).Add(time.Hour * 2),
	})
	if err != nil {
		return nil, fmt.Errorf("get vacations: %w", err)
	}

	dst := make([]Person, 0, len(persons))
	for _, p := range persons {
		if _, there := ooo[p.ID]; there {
			continue
		}
		dst = append(dst, p)
	}

	return dst, nil
}

func (srv *apiServer) serveError(w http.ResponseWriter, err error, msg string) {
	srv.log.Err(err).Msg(msg)
	w.WriteHeader(http.StatusInternalServerError)
}

func (srv *apiServer) serveJSON(w http.ResponseWriter, src interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		srv.log.Err(err).Msg("json marshal")
	}
}
