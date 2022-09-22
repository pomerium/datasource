package gitlab

import "errors"

// Errors
var (
	ErrPrivateTokenRequired = errors.New("gitlab: private token is required")
)
