package wellknownips

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchAzureIPRanges(t *testing.T) {
	t.Parallel()

	ctx, clearTimeout := context.WithTimeout(context.Background(), time.Second*10)
	defer clearTimeout()

	ranges, err := FetchAzureIPRanges(ctx)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Values), 0) {
		assert.Equal(t, ranges.Values[0].ID, "ActionGroup")
	}
}
