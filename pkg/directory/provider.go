package directory

import "context"

// A Provider returns all the groups and users in a directory.
type Provider interface {
	GetDirectory(context.Context) (Bundle, error)
}

// A ProviderFunc implements the Provider interface via a function.
type ProviderFunc func(context.Context) (Bundle, error)

// GetDirectory gets all the groups and users in a directory.
func (p ProviderFunc) GetDirectory(ctx context.Context) (Bundle, error) {
	return p(ctx)
}
