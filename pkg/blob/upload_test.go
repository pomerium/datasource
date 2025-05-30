package blob_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/blob"
)

func TestUploadBundle(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	dir := t.TempDir()
	err := blob.UploadBundle(ctx, "file://"+dir, map[string]any{"a": "x", "b": "y", "c": "z"})
	assert.NoError(t, err)

	bs, err := os.ReadFile(filepath.Join(dir, "bundle.zip"))
	assert.NoError(t, err)

	assert.Equal(t, map[string]any{"a": "x", "b": "y", "c": "z"}, decodeBundle(t, bytes.NewReader(bs)))
}

func decodeBundle(tb testing.TB, r io.Reader) map[string]any {
	tb.Helper()

	bs, err := io.ReadAll(r)
	require.NoError(tb, err)

	zr, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	require.NoError(tb, err)

	m := map[string]any{}
	for _, f := range zr.File {
		r, err := f.Open()
		require.NoError(tb, err)
		var obj any
		require.NoError(tb, json.NewDecoder(r).Decode(&obj))
		m[strings.TrimSuffix(f.Name, ".json")] = obj
	}
	return m
}
