package kv

import (
	pebble "github.com/cockroachdb/pebble/v2"
	"github.com/cockroachdb/pebble/v2/vfs"
)

// NewMemoryStore creates a purely in-memory store.
func NewMemoryStore() Store {
	return NewPebbleStoreWithOptions("", &pebble.Options{
		FS: vfs.NewMem(),
	}, nil, nil)
}
