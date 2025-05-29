package blob_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/internal/blob"
)

func TestDownloadState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	urlstr := "file://" + dir

	assert.NoError(t, blob.UploadState(t.Context(), urlstr, func(dst io.Writer) error {
		_, err := io.WriteString(dst, "STATE")
		return err
	}))
	assert.NoError(t, blob.DownloadState(t.Context(), urlstr, func(src io.Reader) error {
		state, err := io.ReadAll(src)
		if err != nil {
			return err
		}
		assert.Equal(t, "STATE", string(state))
		return nil
	}))
}
