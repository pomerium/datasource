package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/util"
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
			err := util.Remap(tc.src, tc.remap)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dst, tc.src)
		})
	}
}
