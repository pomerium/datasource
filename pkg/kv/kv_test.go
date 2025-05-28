package kv_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/kv"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	s1 := kv.NewMemoryStore()
	s2 := kv.NewMemoryStore()

	assert.NoError(t, s1.Set(t.Context(), []byte("k1"), []byte("v1")))
	assert.NoError(t, s2.Set(t.Context(), []byte("k2"), []byte("v2")))
	assert.NoError(t, s1.Set(t.Context(), []byte("k3"), []byte("v3")))
	assert.NoError(t, s2.Set(t.Context(), []byte("k3"), []byte("v3")))
	assert.NoError(t, s1.Set(t.Context(), []byte("k4"), []byte("v4a")))
	assert.NoError(t, s2.Set(t.Context(), []byte("k4"), []byte("v4b")))
	assert.NoError(t, s1.Set(t.Context(), []byte("k5"), []byte("v5")))

	var diffs []kv.Diff
	for diff, err := range kv.ComputeDiff(t.Context(), s1, s2) {
		assert.NoError(t, err)
		diffs = append(diffs, diff)
	}

	assert.Equal(t, []kv.Diff{
		{[]byte("k1"), []byte("v1"), nil},
		{[]byte("k2"), nil, []byte("v2")},
		{[]byte("k4"), []byte("v4a"), []byte("v4b")},
		{[]byte("k5"), []byte("v5"), nil},
	}, diffs)
}
