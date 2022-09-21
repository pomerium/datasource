// Package okta contains the Okta directory provider.
package okta

import (
	"errors"
)

// Errors.
var (
	ErrAPIKeyRequired           = errors.New("okta: api_key is required")
	ErrServiceAccountNotDefined = errors.New("okta: service account not defined")
	ErrProviderURLNotDefined    = errors.New("okta: provider url not defined")
)

// A ServiceAccount is used by the Okta provider to query the API.
type ServiceAccount struct {
	APIKey string
}
