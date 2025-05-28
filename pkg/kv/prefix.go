package kv

import (
	"bytes"
	"context"
	"iter"
	"slices"
)

type prefixedStore struct {
	underlying Store
	prefix     []byte
}

// NewPrefixedStore takes a store and prepends prefix to the keys for all methods.
func NewPrefixedStore(underlying Store, prefix []byte) Store {
	return &prefixedStore{underlying: underlying, prefix: prefix}
}

func (s *prefixedStore) Delete(ctx context.Context, key []byte) error {
	return s.underlying.Delete(ctx, slices.Concat(s.prefix, key))
}

func (s *prefixedStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	return s.underlying.Get(ctx, slices.Concat(s.prefix, key))
}

func (s *prefixedStore) Iterate(ctx context.Context, prefix []byte) iter.Seq2[Pair, error] {
	return func(yield func(Pair, error) bool) {
		for kvp, err := range s.underlying.Iterate(ctx, slices.Concat(s.prefix, prefix)) {
			if err != nil {
				yield(kvp, err)
				return
			}

			if !yield(Pair{Key: bytes.TrimPrefix(kvp.Key, s.prefix), Value: kvp.Value}, nil) {
				return
			}
		}
	}
}

func (s *prefixedStore) Set(ctx context.Context, key, value []byte) error {
	return s.underlying.Set(ctx, slices.Concat(s.prefix, key), value)
}
