package util_test

import (
	"strings"
	"testing"

	"github.com/pomerium/datasource/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
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
		t.Run(strings.Join(tc.tags, ","), func(t *testing.T) {
			assert.Equal(t, tc.tags, util.GetStructTagNames(tc.in, "tag"))
		})
	}
}
