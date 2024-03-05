package wellknownips

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchAmazonAWSIPRanges(t *testing.T) {
	t.Parallel()

	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	defer clearTimeout()

	client := http.DefaultClient

	ranges, err := FetchAmazonAWSIPRanges(ctx, client, DefaultAmazonAWSIPRangesURL)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Prefixes), 0) {
		assert.Equal(t, "3.5.140.0/22", ranges.Prefixes[0].IPPrefix)
	}
}
