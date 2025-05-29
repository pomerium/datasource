package kv

import (
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

func (s *pebbleStore) All(ctx context.Context) iter.Seq2[Pair, error] {
	return s.iterate(ctx, s.iterOptions)
}

func (s *pebbleStore) AllPrefix(ctx context.Context, prefix []byte) iter.Seq2[Pair, error] {
	iterOptions := new(pebble.IterOptions)
	if s.iterOptions != nil {
		*iterOptions = *s.iterOptions
	}
	iterOptions.LowerBound = prefix
	iterOptions.UpperBound = prefixToUpperBound(prefix)
	return s.iterate(ctx, iterOptions)
}

func (s *pebbleStore) Delete(_ context.Context, key []byte) error {
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

func (s *pebbleStore) DeletePrefix(_ context.Context, prefix []byte) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}

	err = db.DeleteRange(prefix, prefixToUpperBound(prefix), s.writeOptions)
	if err != nil {
		return fmt.Errorf("pebble: error deleting prefix: %w", err)
	}

	return nil
}

func (s *pebbleStore) Get(_ context.Context, key []byte) ([]byte, error) {
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

func (s *pebbleStore) Set(_ context.Context, key, value []byte) error {
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

func (s *pebbleStore) getDB() (*pebble.DB, error) {
	s.dbOnce.Do(func() {
		s.db, s.dbErr = pebble.Open(s.dirname, s.options)
		if s.dbErr != nil {
			s.dbErr = fmt.Errorf("pebble: error opening database: %w", s.dbErr)
		}
	})
	return s.db, s.dbErr
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
