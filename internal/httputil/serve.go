package httputil

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"sort"
	"time"

	"golang.org/x/exp/maps"
)

// EncodeBundle encodes a bundle to a writer.
func EncodeBundle(w io.Writer, bundle map[string]any) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	recordTypes := maps.Keys(bundle)
	sort.Strings(recordTypes)

	for _, recordType := range recordTypes {
		fw, err := zw.Create(recordType + ".json")
		if err != nil {
			return fmt.Errorf("failed to create %s file: %w", recordType, err)
		}
		err = json.NewEncoder(fw).Encode(bundle[recordType])
		if err != nil {
			return fmt.Errorf("failed to write %s data: %w", recordType, err)
		}
	}

	err := zw.Close()
	if err != nil {
		return fmt.Errorf("failed to close zip file: %w", err)
	}

	return nil
}

// ServeBundle serves a bundle of data.
func ServeBundle(w http.ResponseWriter, r *http.Request, bundle map[string]any) error {
	var buf bytes.Buffer
	err := EncodeBundle(&buf, bundle)
	if err != nil {
		return fmt.Errorf("failed to encode bundle: %w", err)
	}

	w.Header().Set("Content-Type", "application/zip")
	return ServeData(w, r, "bundle.zip", buf.Bytes())
}

// ServeContent serves content over http.
func ServeContent(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	hash uint64,
	content io.ReadSeeker,
) error {
	w.Header().Set("ETag", fmt.Sprintf(`"%x"`, hash))
	http.ServeContent(w, r, name, time.Now(), content)
	return nil
}

// ServeData serves data over http.
func ServeData(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	data []byte,
) error {
	hasher := fnv.New64()
	_, _ = hasher.Write(data)
	h := hasher.Sum64()
	return ServeContent(w, r, name, h, bytes.NewReader(data))
}
