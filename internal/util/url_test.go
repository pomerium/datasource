package util_test

import (
	"encoding/json"
	"testing"

	"github.com/pomerium/datasource/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURL(t *testing.T) {
	for _, tc := range []struct {
		JSON string
	}{
		{`{"url":"http://company.com/path?query"}`},
		{`{"url":null}`},
	} {
		t.Run(tc.JSON, func(t *testing.T) {
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
