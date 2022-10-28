package ip2location

import (
	"bytes"
	"net/http"

	"github.com/pomerium/datasource/internal/httputil"
	"github.com/pomerium/datasource/internal/jsonutil"
)

type serverConfig struct {
	file string
}

// A ServerOption customizes the server config.
type ServerOption func(*serverConfig)

// WithFile sets the file for the config.
func WithFile(file string) ServerOption {
	return func(cfg *serverConfig) {
		cfg.file = file
	}
}

func getServerConfig(options ...ServerOption) *serverConfig {
	cfg := new(serverConfig)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

// Server serves ip2location records
type Server struct {
	cfg *serverConfig
}

// NewServer creates a new Server.
func NewServer(options ...ServerOption) *Server {
	cfg := getServerConfig(options...)
	return &Server{
		cfg: cfg,
	}
}

// ServeHTTP implements the http.Handler interface.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := srv.serveHTTP(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv *Server) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	var buf bytes.Buffer
	dst := jsonutil.NewJSONArrayStream(&buf)
	err := fileToJSON(dst, srv.cfg.file)
	if err != nil {
		return err
	}
	return httputil.ServeData(w, r, "ip2location.json", buf.Bytes())
}
