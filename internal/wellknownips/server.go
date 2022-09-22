package wellknownips

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

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

	eg, ctx := errgroup.WithContext(r.Context())
	recordLookup := map[string][]Record{}

	recordLookup[AmazonASNumber] = nil
	eg.Go(func() error {
		amazonAWSIPRanges, err := FetchAmazonAWSIPRanges(ctx, transport.Client(), DefaultAmazonAWSIPRangesURL)
		if err != nil {
			return fmt.Errorf("error fetching amazon aws ip ranges: %w", err)
		}
		recordLookup[AmazonASNumber] = RecordsFromAmazonAWSIPRanges(amazonAWSIPRanges)
		return nil
	})

	recordLookup[AppleASNumber] = nil
	eg.Go(func() error {
		recordLookup[AppleASNumber] = RecordsFromAppleDomainVerificationIPAddresses(AppleDomainVerificationIPAddresses)
		return nil
	})

	recordLookup[AtlassianASNumber] = nil
	eg.Go(func() error {
		atlassianRanges, err := FetchAtlassianIPRanges(ctx, transport.Client(), DefaultAtlassianIPRangesURL)
		if err != nil {
			return fmt.Errorf("error fetching atlassian ip ranges: %w", err)
		}
		recordLookup[AtlassianASNumber] = RecordsFromAtlassianIPRanges(atlassianRanges)
		return nil
	})

	recordLookup[MicrosoftASNumber] = nil
	eg.Go(func() error {
		azureRanges, err := FetchAzureIPRanges(ctx)
		if err != nil {
			return fmt.Errorf("error fetching azure ip ranges: %w", err)
		}
		recordLookup[MicrosoftASNumber] = RecordsFromAzureIPRanges(azureRanges)
		return err
	})

	recordLookup[GitHubASNumber] = nil
	eg.Go(func() error {
		githubMeta, err := FetchGitHubMeta(ctx, transport.Client(), DefaultGitHubMetaURL)
		if err != nil {
			return fmt.Errorf("error fetching github ip ranges: %w", err)
		}
		recordLookup[GitHubASNumber] = RecordsFromGitHubMeta(githubMeta)
		return err
	})

	recordLookup[StripeASNumber] = nil
	eg.Go(func() error {
		stripeRanges, err := FetchStripeIPRanges(ctx, transport.Client(), DefaultStripeIPRangesURL)
		if err != nil {
			return fmt.Errorf("error fetching stripe ip ranges: %w", err)
		}
		recordLookup[StripeASNumber] = RecordsFromStripeIPRanges(stripeRanges)
		return err
	})

	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("error fetching well known ip ranges: %w", err)
	}

	stream, err := FetchIP2ASNDatabase(r.Context(), transport.Client(), srv.cfg.ip2asnURL)
	if err != nil {
		return fmt.Errorf("error fetching ip2asn database: %w", err)
	}

	dst := jsonutil.NewJSONArrayStream(w)
	for stream.Next(r.Context()) {
		_, ok := recordLookup[stream.Record().ASNumber]
		if ok {
			// skip well-defined ip ranges
			continue
		}

		for _, record := range RecordsFromIP2ASNRecord(stream.Record()) {
			err := dst.Encode(record)
			if err != nil {
				return fmt.Errorf("failed to write record to destination: %w", err)
			}
		}
	}

	var keys []string
	for key := range recordLookup {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		for _, record := range recordLookup[key] {
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
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			srv.cacheErr = fmt.Errorf("failed to create wellknownips directory")
			return
		}

		srv.cache = diskcache.New(dir)
	})
	return srv.cache, srv.cacheErr
}
