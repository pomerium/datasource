package fleetdm

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pomerium/datasource/internal/fleetdm/client"
)

func NewServer(opts ...Option) (*mux.Router, error) {
	cfg := newConfig(opts...)

	client, err := client.New(client.WithToken(cfg.apiToken), client.WithURL(cfg.apiURL))
	if err != nil {
		return nil, err
	}

	srv := server{
		cfg:    cfg,
		client: client,
	}

	r := mux.NewRouter()
	r.Path("/").Methods(http.MethodGet).HandlerFunc(srv.getIndexHandler)

	return r, nil
}

type server struct {
	cfg    *config
	client *client.Client
}
