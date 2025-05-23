package wellknownips

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchGitHubMeta(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	client := http.DefaultClient

	meta, err := FetchGitHubMeta(ctx, client, DefaultGitHubMetaURL)
	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Contains(t, meta.Hooks, "192.30.252.0/22")
}
