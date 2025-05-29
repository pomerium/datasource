package kv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"sync"

	pebble "github.com/cockroachdb/pebble/v2"
)

type pebbleStore struct {
	dirname      string
	options      *pebble.Options
	writeOptions *pebble.WriteOptions
	iterOptions  *pebble.IterOptions

	dbOnce sync.Once
	db     *pebble.DB
	dbErr  error
}

// NewPebbleStore creates a new store using pebble.
func NewPebbleStore(dirname string) Store {
	return NewPebbleStoreWithOptions(dirname, nil, nil, nil)
}

// NewPebbleStoreWithOptions creates a new store using pebble with the given options.
func NewPebbleStoreWithOptions(
	dirname string,
	options *pebble.Options,
	writeOptions *pebble.WriteOptions,
	iterOptions *pebble.IterOptions,
) Store {
	return &pebbleStore{
		dirname:      dirname,
		options:      options,
		writeOptions: writeOptions,
		iterOptions:  iterOptions,
	}
}

func (s *pebbleStore) Delete(_ context.Context, key []byte) error {
	return s.delete(key)
}

func (s *pebbleStore) DeleteAll(_ context.Context) error {
	return s.deletePrefix(nil)
}

func (s *pebbleStore) Get(_ context.Context, key []byte) ([]byte, error) {
	return s.get(key)
}

func (s *pebbleStore) IterateAll(ctx context.Context) iter.Seq2[Pair, error] {
	return s.iterate(ctx, s.iterOptions)
}

func (s *pebbleStore) Prefix(prefix []byte) Store {
	return &pebblePrefixStore{underlying: s, prefix: prefix}
}

func (s *pebbleStore) Set(_ context.Context, key, value []byte) error {
	return s.set(key, value)
}

func (s *pebbleStore) get(key []byte) ([]byte, error) {
	db, err := s.getDB()
	if err != nil {
		return nil, err
	}

	value, closer, err := db.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("pebble: error getting key: %w", err)
	}
	value = slices.Clone(value)
	_ = closer.Close()

	return value, nil
}

func (s *pebbleStore) getDB() (*pebble.DB, error) {
	s.dbOnce.Do(func() {
		options := new(pebble.Options)
		if s.options != nil {
			*options = *s.options
		}
		options.LoggerAndTracer = pebbleLogger{}

		s.db, s.dbErr = pebble.Open(s.dirname, options)
		if s.dbErr != nil {
			s.dbErr = fmt.Errorf("pebble: error opening database: %w", s.dbErr)
		}
	})
	return s.db, s.dbErr
}

func (s *pebbleStore) delete(key []byte) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}

	err = db.Delete(key, s.writeOptions)
	if err != nil {
		return fmt.Errorf("pebble: error deleting key: %w", err)
	}

	return nil
}

func (s *pebbleStore) deletePrefix(prefix []byte) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}

	err = db.DeleteRange(prefix, prefixToUpperBound(prefix), s.writeOptions)
	if err != nil {
		return fmt.Errorf("pebble: error deleting keys: %w", err)
	}

	return nil
}

func (s *pebbleStore) iterate(ctx context.Context, iterOptions *pebble.IterOptions) iter.Seq2[Pair, error] {
	return func(yield func(Pair, error) bool) {
		db, err := s.getDB()
		if err != nil {
			yield(Pair{}, err)
			return
		}

		it, err := db.NewIterWithContext(ctx, iterOptions)
		if err != nil {
			yield(Pair{}, fmt.Errorf("pebble: error creating iterator: %w", err))
			return
		}

		for it.First(); it.Valid(); it.Next() {
			pair := Pair{slices.Clone(it.Key()), slices.Clone(it.Value())}
			if !yield(pair, nil) {
				_ = it.Close()
				return
			}
		}

		err = it.Error()
		if err != nil {
			_ = it.Close()
			yield(Pair{}, fmt.Errorf("pebble: error iterating over key value pairs: %w", err))
			return
		}

		err = it.Close()
		if err != nil {
			yield(Pair{}, fmt.Errorf("pebble: error closing iterator: %w", err))
			return
		}
	}
}

func (s *pebbleStore) set(key, value []byte) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}

	err = db.Set(key, value, s.writeOptions)
	if err != nil {
		return fmt.Errorf("pebble: error setting key: %w", err)
	}

	return nil
}

type pebblePrefixStore struct {
	underlying *pebbleStore
	prefix     []byte
}

func (s *pebblePrefixStore) Delete(_ context.Context, key []byte) error {
	return s.underlying.delete(slices.Concat(s.prefix, key))
}

func (s *pebblePrefixStore) DeleteAll(_ context.Context) error {
	return s.underlying.deletePrefix(s.prefix)
}

func (s *pebblePrefixStore) Get(_ context.Context, key []byte) ([]byte, error) {
	return s.underlying.get(slices.Concat(s.prefix, key))
}

func (s *pebblePrefixStore) IterateAll(ctx context.Context) iter.Seq2[Pair, error] {
	return func(yield func(Pair, error) bool) {
		iterOptions := new(pebble.IterOptions)
		if s.underlying.iterOptions != nil {
			*iterOptions = *s.underlying.iterOptions
		}
		iterOptions.LowerBound = s.prefix
		iterOptions.UpperBound = prefixToUpperBound(s.prefix)

		for pair, err := range s.underlying.iterate(ctx, iterOptions) {
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

func (s *pebblePrefixStore) Prefix(prefix []byte) Store {
	return &pebblePrefixStore{underlying: s.underlying, prefix: slices.Concat(s.prefix, prefix)}
}

func (s *pebblePrefixStore) Set(_ context.Context, key, value []byte) error {
	return s.underlying.set(slices.Concat(s.prefix, key), value)
}

func prefixToUpperBound(prefix []byte) []byte {
	upperBound := make([]byte, len(prefix))
	copy(upperBound, prefix)
	for i := len(upperBound) - 1; i >= 0; i-- {
		upperBound[i] = upperBound[i] + 1
		if upperBound[i] != 0 {
			return upperBound[:i+1]
		}
	}
	return nil // no upper-bound
}

type pebbleLogger struct{}

func (pebbleLogger) Infof(_ string, _ ...any)                     {}
func (pebbleLogger) Errorf(_ string, _ ...any)                    {}
func (pebbleLogger) Fatalf(_ string, _ ...any)                    {}
func (pebbleLogger) Eventf(_ context.Context, _ string, _ ...any) {}
func (pebbleLogger) IsTracingEnabled(_ context.Context) bool      { return false }
