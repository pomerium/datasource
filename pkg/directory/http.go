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
		err := h.serve(r.Context(), w)
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

func (h *handler) serve(ctx context.Context, w http.ResponseWriter) error {
	groups, users, err := h.provider.GetDirectory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get directory data: %w", err)
	}
	w.Header().Set("Content-Type", "application/zip")
	return encodeBundle(w, groups, users)
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
		{"pomerium.io/DirectoryGroup.json", &groups},
		{"pomerium.io/DirectoryUser.json", &users},
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

func encodeBundle(w io.Writer, groups []Group, users []User) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, file := range []struct {
		name string
		data any
	}{
		{"pomerium.io/DirectoryGroup.json", groups},
		{"pomerium.io/DirectoryUser.json", users},
	} {
		fw, err := zw.Create(file.name)
		if err != nil {
			return fmt.Errorf("failed to create %s file: %w", file.name, err)
		}
		err = json.NewEncoder(fw).Encode(file.data)
		if err != nil {
			return fmt.Errorf("failed to write %s data: %w", file.name, err)
		}
	}

	err := zw.Close()
	if err != nil {
		return fmt.Errorf("failed to close zip file: %w", err)
	}

	return nil
}
