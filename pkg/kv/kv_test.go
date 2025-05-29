package kv_test

import (
	"iter"
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
	for diff, err := range kv.ComputeDiff(s1.IterateAll(t.Context()), s2.IterateAll(t.Context())) {
		assert.NoError(t, err)
		diffs = append(diffs, diff)
	}

	assert.Equal(t, []kv.Diff{
		{[]byte("k1"), []byte("v1"), nil},
		{[]byte("k2"), nil, []byte("v2")},
		{[]byte("k4"), []byte("v4a"), []byte("v4b")},
		{[]byte("k5"), []byte("v5"), nil},
	}, diffs)

	// test prefixing

	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k1"), []byte("v1")))
	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k2"), []byte("v2")))
	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k3"), []byte("v3")))
	s3 := s1.Prefix([]byte("prefix-"))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}, collectPairs(t, s3.IterateAll(t.Context())))

	assert.NoError(t, s3.Set(t.Context(), []byte("k4"), []byte("v4")))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
		{"k4", "v4"},
	}, collectPairs(t, s3.IterateAll(t.Context())))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k3", "v3"},
		{"k4", "v4a"},
		{"k5", "v5"},
		{"prefix-k1", "v1"},
		{"prefix-k2", "v2"},
		{"prefix-k3", "v3"},
		{"prefix-k4", "v4"},
	}, collectPairs(t, s1.IterateAll(t.Context())))

	assert.NoError(t, s3.Delete(t.Context(), []byte("k4")))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}, collectPairs(t, s3.IterateAll(t.Context())))

	v, err := s3.Get(t.Context(), []byte("k1"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)
}

func collectPairs(tb testing.TB, seq iter.Seq2[kv.Pair, error]) [][2]string {
	tb.Helper()

	var pairs [][2]string
	for pair, err := range seq {
		if assert.NoError(tb, err) {
			pairs = append(pairs, [2]string{string(pair[0]), string(pair[1])})
		}
	}
	return pairs
}
