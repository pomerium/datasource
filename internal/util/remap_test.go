package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/util"
)

func TestRemap(t *testing.T) {
	for name, tc := range map[string]struct {
		remap    []util.FieldRemap
		src, dst []map[string]interface{}
	}{
		"none": {
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
		"simple": {
			remap: []util.FieldRemap{
				{"a", "k"},
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
		"overlapping": {
			remap: []util.FieldRemap{
				{"id", "new_id"},
				{"email", "id"},
			},
			src: []map[string]interface{}{{
				"id":    1,
				"email": "me@corp.com",
			}},
			dst: []map[string]interface{}{{
				"new_id": 1,
				"id":     "me@corp.com",
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
