package kv_test

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/kv"
)

func TestPrefixedStore(t *testing.T) {
	t.Parallel()

	s1 := kv.NewMemoryStore()
	assert.NoError(t, s1.Set(t.Context(), []byte("other-k1"), []byte("v1")))
	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k1"), []byte("v1")))
	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k2"), []byte("v2")))
	assert.NoError(t, s1.Set(t.Context(), []byte("prefix-k3"), []byte("v3")))

	s2 := kv.NewPrefixedStore(s1, []byte("prefix-"))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}, collectPairs(t, s2.All(t.Context())))

	assert.NoError(t, s2.Set(t.Context(), []byte("k4"), []byte("v4")))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
		{"k4", "v4"},
	}, collectPairs(t, s2.All(t.Context())))
	assert.Equal(t, [][2]string{
		{"other-k1", "v1"},
		{"prefix-k1", "v1"},
		{"prefix-k2", "v2"},
		{"prefix-k3", "v3"},
		{"prefix-k4", "v4"},
	}, collectPairs(t, s1.All(t.Context())))

	assert.NoError(t, s2.Delete(t.Context(), []byte("k4")))
	assert.Equal(t, [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"k3", "v3"},
	}, collectPairs(t, s2.All(t.Context())))

	v, err := s2.Get(t.Context(), []byte("k1"))
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
