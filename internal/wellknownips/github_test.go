package wellknownips

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchGitHubMeta(t *testing.T) {
	t.Parallel()

	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	defer clearTimeout()

	client := http.DefaultClient

	meta, err := FetchGitHubMeta(ctx, client, DefaultGitHubMetaURL)
	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Contains(t, meta.Hooks, "192.30.252.0/22")
}
