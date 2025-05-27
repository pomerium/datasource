package directory

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pomerium/datasource/internal/httputil"
)

type handler struct {
	router   *chi.Mux
	provider Provider
}

// NewHandler creates a new Handler.
func NewHandler(provider Provider) http.Handler {
	h := &handler{provider: provider}
	h.router = chi.NewMux()
	h.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		err := h.serve(r.Context(), w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	return h
}

// ServeHTTP serves an HTTP request with directory users and groups.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *handler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bundle, err := h.provider.GetDirectory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get directory data: %w", err)
	}

	return httputil.ServeBundle(w, r, bundle)
}

func decodeBundle(r io.Reader) (groups []Group, users []User, err error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open zip file for reading: %w", err)
	}

	for _, file := range []struct {
		name string
		ptr  any
	}{
		{GroupRecordType + ".json", &groups},
		{UserRecordType + ".json", &users},
	} {
		fr, err := zr.Open(file.name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open file %s: %w", file.name, err)
		}

		err = json.NewDecoder(fr).Decode(file.ptr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode file %s: %w", file.name, err)
		}

		err = fr.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to close file %s: %w", file.name, err)
		}
	}
	return groups, users, nil
}
