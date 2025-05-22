package blob

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// NewHandler creates a new HTTP handler for blob storage.
func NewHandler(urlstr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket, err := openBucket(r.Context(), urlstr)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("error opening bucket")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer bucket.Close()

		file, err := bucket.NewReader(r.Context(), "bundle.zip", nil)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("error serving file from bucket")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		http.ServeContent(w, r, "bundle.zip", file.ModTime(), file)
	})
}
