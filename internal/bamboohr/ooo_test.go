package bamboohr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOOO(t *testing.T) {
	t.Parallel()

	now := time.Now()
	for _, tc := range []struct {
		name string
		ooo  bool
		out  []Period
	}{
		{"empty", false, nil},
		{"simple", true, []Period{
			{Start: now.Add(-time.Second), End: now.Add(time.Second)},
		}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.ooo, isOut(now, tc.out))
		})
	}
}
