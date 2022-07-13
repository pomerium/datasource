package wellknownips

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchAzureIPRanges(t *testing.T) {
	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	defer clearTimeout()

	client := http.DefaultClient

	ranges, err := FetchAzureIPRanges(ctx, client, DefaultAzureIPRangesURL)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Values), 0) {
		assert.Equal(t, ranges.Values[0].ID, "ActionGroup")
	}
}
