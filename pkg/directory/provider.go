package directory

import "context"

// A Provider returns all the groups and users in a directory.
type Provider interface {
	GetDirectory(context.Context) ([]Group, []User, error)
}

// A ProviderFunc implements the Provider interface via a function.
type ProviderFunc func(context.Context) ([]Group, []User, error)

// GetDirectory gets all the groups and users in a directory.
func (p ProviderFunc) GetDirectory(ctx context.Context) ([]Group, []User, error) {
	return p(ctx)
}

// A PersistentProvider is a directory provider that supports getting and setting directory state.
type PersistentProvider interface {
	Provider
	GetDirectoryState(ctx context.Context) ([]byte, error)
	SetDirectoryState(ctx context.Context, state []byte) error
}
