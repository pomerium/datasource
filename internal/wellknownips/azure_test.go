package wellknownips

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchAzureIPRanges(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	ranges, err := FetchAzureIPRanges(ctx)
	assert.NoError(t, err)
	if assert.NotNil(t, ranges) && assert.Greater(t, len(ranges.Values), 0) {
		assert.Equal(t, ranges.Values[0].ID, "ActionGroup")
	}
}
