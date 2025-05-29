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

func (s *prefixedStore) All(ctx context.Context) iter.Seq2[Pair, error] {
	return s.AllPrefix(ctx, nil)
}

func (s *prefixedStore) AllPrefix(ctx context.Context, prefix []byte) iter.Seq2[Pair, error] {
	return func(yield func(Pair, error) bool) {
		for pair, err := range s.underlying.AllPrefix(ctx, slices.Concat(s.prefix, prefix)) {
			if err != nil {
				yield(pair, err)
				return
			}

			if !yield(Pair{bytes.TrimPrefix(pair[0], s.prefix), pair[1]}, nil) {
				return
			}
		}
	}
}

func (s *prefixedStore) Delete(ctx context.Context, key []byte) error {
	return s.underlying.Delete(ctx, slices.Concat(s.prefix, key))
}

func (s *prefixedStore) DeletePrefix(ctx context.Context, prefix []byte) error {
	return s.underlying.DeletePrefix(ctx, slices.Concat(s.prefix, prefix))
}

func (s *prefixedStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	return s.underlying.Get(ctx, slices.Concat(s.prefix, key))
}

func (s *prefixedStore) Set(ctx context.Context, key, value []byte) error {
	return s.underlying.Set(ctx, slices.Concat(s.prefix, key), value)
}
