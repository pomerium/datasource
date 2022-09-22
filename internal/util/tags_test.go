package util_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/internal/util"
)

func TestTags(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   any
		tags []string
	}{
		{
			struct{}{},
			[]string{},
		},
		{
			struct {
				A string `json:"xx" tag:"one"`
				B string `tag:"two"`
				C string `json:"three"`
				D string `json:"dd" tag:",omitempty"`
				E string `json:"ee" tag:"four,omitempty"`
				F string `tag:"-,omitempty"`
				G string `tag:"-"`
			}{},
			[]string{"one", "two", "four"},
		},
	} {
		tc := tc
		t.Run(strings.Join(tc.tags, ","), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.tags, util.GetStructTagNames(tc.in, "tag"))
		})
	}
}
