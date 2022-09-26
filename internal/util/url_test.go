package util_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/util"
)

func TestURL(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		JSON string
	}{
		{`{"url":"http://company.com/path?query"}`},
		{`{"url":null}`},
	} {
		tc := tc
		t.Run(tc.JSON, func(t *testing.T) {
			t.Parallel()

			var in struct {
				URL *util.URL `json:"url"`
			}
			require.NoError(t, json.Unmarshal([]byte(tc.JSON), &in))
			data, err := json.Marshal(&in)
			require.NoError(t, err)
			assert.Equal(t, tc.JSON, string(data))
		})
	}
}
