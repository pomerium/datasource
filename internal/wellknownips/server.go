package wellknownips

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/jsonutil"
)

var DefaultIP2ASNURL = "https://iptoasn.com/data/ip2asn-v4.tsv.gz"

type serverConfig struct {
	ip2asnURL string
}

// A ServerOption customizes the server config.
type ServerOption func(*serverConfig)

// WithIP2ASNURL sets the ip2asn url in the config.
func WithIP2ASNURL(url string) ServerOption {
	return func(cfg *serverConfig) {
		cfg.ip2asnURL = url
	}
}

func getServerConfig(options ...ServerOption) *serverConfig {
	cfg := new(serverConfig)
	WithIP2ASNURL(DefaultIP2ASNURL)(cfg)
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

// Server serves well-known-ip records
type Server struct {
	cfg *serverConfig

	cacheInit sync.Once
	cache     httpcache.Cache
	cacheErr  error
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
		log.Error().Err(err).Send()
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv *Server) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return nil
	}

	cache, err := srv.getCache()
	if err != nil {
		return fmt.Errorf("error getting cache: %w", err)
	}

	transport := httpcache.NewTransport(cache)
	stream, err := FetchIP2ASNDatabase(r.Context(), transport.Client(), srv.cfg.ip2asnURL)
	if err != nil {
		return fmt.Errorf("error fetching ip2asn database: %w", err)
	}

	dst := jsonutil.NewJSONArrayStream(w)
	for stream.Next(r.Context()) {
		for _, record := range RecordsFromIP2ASNRecord(stream.Record()) {
			err := dst.Encode(record)
			if err != nil {
				return fmt.Errorf("failed to write record to destination: %w", err)
			}
		}
	}

	err = stream.Err()
	if err != nil {
		return err
	}

	err = dst.Close()
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) getCache() (httpcache.Cache, error) {
	srv.cacheInit.Do(func() {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			srv.cacheErr = fmt.Errorf("failed to get user cache directory: %w", err)
			return
		}

		dir := filepath.Join(userCacheDir, "pomerium-datasource", "wellknownips")
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			srv.cacheErr = fmt.Errorf("failed to create wellknownips directory")
			return
		}

		srv.cache = diskcache.New(dir)
	})
	return srv.cache, srv.cacheErr
}
