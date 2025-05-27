package blob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/internal/blob"
)

func TestDownloadState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	urlstr := "file://" + dir

	assert.NoError(t, blob.UploadState(t.Context(), urlstr, []byte("STATE")))
	state, err := blob.DownloadState(t.Context(), urlstr)
	assert.NoError(t, err)
	assert.Equal(t, "STATE", string(state))
}
