package internal_test

import (
	"testing"

	"github.com/pomerium/datasource/bamboohr/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemap(t *testing.T) {
	for name, tc := range map[string]struct {
		remap    map[string]string
		src, dst []map[string]interface{}
	}{
		"no remap": {
			remap: nil,
			src: []map[string]interface{}{{
				"a": 1,
				"b": 2,
			}},
			dst: []map[string]interface{}{{
				"b": 2,
				"a": 1,
			}},
		},
		"simple remap": {
			remap: map[string]string{
				"a": "k",
			},
			src: []map[string]interface{}{{
				"a": 1,
				"b": 2,
			}},
			dst: []map[string]interface{}{{
				"b": 2,
				"k": 1,
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := internal.Remap(tc.src, tc.remap)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dst, tc.src)
		})
	}
}
