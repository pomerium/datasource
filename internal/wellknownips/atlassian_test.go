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
		assert.Equal(t, "52.82.172.0/22", ranges.Items[0].CIDR)
	}
}
