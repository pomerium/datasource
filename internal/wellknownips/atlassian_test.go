package wellknownips

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchAtlassianIPRanges(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	client := http.DefaultClient

	ranges, err := FetchAtlassianIPRanges(ctx, client, DefaultAtlassianIPRangesURL)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Items), 0) {
		assert.Equal(t, ranges.Items[0].CIDR, "23.249.208.0/20")
	}
}
