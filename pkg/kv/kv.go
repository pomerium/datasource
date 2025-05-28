package kv

import (
	"bytes"
	"context"
	"errors"
	"iter"
)

// ErrNotFound indicates a key was not found in the store.
var ErrNotFound = errors.New("key not found")

// A Pair is a key and value pair.
type Pair struct {
	Key, Value []byte
}

// An Iterator supports iterating over keys and values in sorted order.
type Iterator interface {
	// Iterate iterates over the keys and values in sorted order.
	Iterate(ctx context.Context) iter.Seq2[Pair, error]
}

// A Store stores key and values in sorted order.
type Store interface {
	Iterator
	// Delete deletes the key from the store.
	Delete(ctx context.Context, key []byte) error
	// Get gets a value for the given key. If not found ErrNotFound will be returned.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Set sets a key value pair in the store.
	Set(ctx context.Context, key, value []byte) error
}

// A Diff is a difference between two key value pairs.
type Diff struct {
	Key, LeftValue, RightValue []byte
}

// ComputeDiff computes the difference between two sorted key value iterators.
func ComputeDiff(ctx context.Context, left, right Iterator) iter.Seq2[Diff, error] {
	return func(yield func(Diff, error) bool) {
		leftNext, leftStop := iter.Pull2(left.Iterate(ctx))
		defer leftStop()

		rightNext, rightStop := iter.Pull2(right.Iterate(ctx))
		defer rightStop()

		leftKVP, leftErr, leftValid := leftNext()
		rightKVP, rightErr, rightValid := rightNext()

		for {
			if leftErr != nil {
				yield(Diff{}, leftErr)
				return
			}
			if rightErr != nil {
				yield(Diff{}, rightErr)
				return
			}

			// if both are invalid, we are done
			if !leftValid && !rightValid {
				return
			}

			// if only the left is valid, yield a left-only key value pair
			if !rightValid {
				if !yield(Diff{Key: leftKVP.Key, LeftValue: leftKVP.Value}, nil) {
					return
				}
				leftKVP, leftErr, leftValid = leftNext()
				continue
			}

			// if only the right is valid, yield a right-only key value pair
			if !leftValid {
				if !yield(Diff{Key: rightKVP.Key, RightValue: rightKVP.Value}, nil) {
					return
				}
				rightKVP, rightErr, rightValid = rightNext()
				continue
			}

			// compare the keys
			c := bytes.Compare(leftKVP.Key, rightKVP.Key)

			// if the left key is less than the right key, yield a left-only key value pair
			if c < 0 {
				if !yield(Diff{Key: leftKVP.Key, LeftValue: leftKVP.Value}, nil) {
					return
				}
				leftKVP, leftErr, leftValid = leftNext()
				continue
			}

			// if the right key is less than the left key, yield a right-only key value pair
			if c > 0 {
				if !yield(Diff{Key: rightKVP.Key, RightValue: rightKVP.Value}, nil) {
					return
				}
				rightKVP, rightErr, rightValid = rightNext()
				continue
			}

			// keys are the same, compare the values, if they are not the same, yield a diff
			if !bytes.Equal(leftKVP.Value, rightKVP.Value) {
				if !yield(Diff{Key: rightKVP.Key, LeftValue: leftKVP.Value, RightValue: rightKVP.Value}, nil) {
					return
				}
			}

			// move left and right forward
			leftKVP, leftErr, leftValid = leftNext()
			rightKVP, rightErr, rightValid = rightNext()
		}
	}
}
