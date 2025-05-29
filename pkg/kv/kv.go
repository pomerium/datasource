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
type Pair = [2][]byte

// A Store stores key and values in sorted order.
type Store interface {
	// Delete deletes the key from the store.
	Delete(ctx context.Context, key []byte) error
	// DeleteAll deletes all the keys from the store.
	DeleteAll(ctx context.Context) error
	// IterateAll iterates over all the keys and values in sorted order.
	IterateAll(ctx context.Context) iter.Seq2[Pair, error]
	// Get gets a value for the given key. If not found ErrNotFound will be returned.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Prefix returns a new store with the given prefix prepended to every key.
	Prefix(prefix []byte) Store
	// Set sets a key value pair in the store.
	Set(ctx context.Context, key, value []byte) error
}

// A Diff is a difference between two key value pairs.
type Diff [3][]byte

// Key returns the diff key.
func (diff Diff) Key() []byte {
	return diff[0]
}

// LeftValue returns the diff left value.
func (diff Diff) LeftValue() []byte {
	return diff[1]
}

// RightValue returns the diff right value.
func (diff Diff) RightValue() []byte {
	return diff[2]
}

// ComputeDiff computes the difference between two sorted key value iterators.
func ComputeDiff(left, right iter.Seq2[Pair, error]) iter.Seq2[Diff, error] {
	return func(yield func(Diff, error) bool) {
		leftNext, leftStop := iter.Pull2(left)
		defer leftStop()

		rightNext, rightStop := iter.Pull2(right)
		defer rightStop()

		leftPair, leftErr, leftValid := leftNext()
		rightPair, rightErr, rightValid := rightNext()

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
				if !yield(Diff{leftPair[0], leftPair[1], nil}, nil) {
					return
				}
				leftPair, leftErr, leftValid = leftNext()
				continue
			}

			// if only the right is valid, yield a right-only key value pair
			if !leftValid {
				if !yield(Diff{rightPair[0], nil, rightPair[1]}, nil) {
					return
				}
				rightPair, rightErr, rightValid = rightNext()
				continue
			}

			// compare the keys
			c := bytes.Compare(leftPair[0], rightPair[0])

			// if the left key is less than the right key, yield a left-only key value pair
			if c < 0 {
				if !yield(Diff{leftPair[0], leftPair[1], nil}, nil) {
					return
				}
				leftPair, leftErr, leftValid = leftNext()
				continue
			}

			// if the right key is less than the left key, yield a right-only key value pair
			if c > 0 {
				if !yield(Diff{rightPair[0], nil, rightPair[1]}, nil) {
					return
				}
				rightPair, rightErr, rightValid = rightNext()
				continue
			}

			// keys are the same, compare the values, if they are not the same, yield a diff
			if !bytes.Equal(leftPair[1], rightPair[1]) {
				if !yield(Diff{rightPair[0], leftPair[1], rightPair[1]}, nil) {
					return
				}
			}

			// move left and right forward
			leftPair, leftErr, leftValid = leftNext()
			rightPair, rightErr, rightValid = rightNext()
		}
	}
}
