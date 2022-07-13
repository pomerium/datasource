package wellknownips

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchAmazonAWSIPRanges(t *testing.T) {
	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	defer clearTimeout()

	client := http.DefaultClient

	ranges, err := FetchAmazonAWSIPRanges(ctx, client, DefaultAmazonAWSIPRangesURL)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Prefixes), 0) {
		assert.Equal(t, ranges.Prefixes[0].IPPrefix, "3.5.140.0/22")
	}
}
