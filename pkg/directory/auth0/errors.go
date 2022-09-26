package auth0

import "errors"

// Errors
var (
	ErrClientIDRequired     = errors.New("auth0: client id is required")
	ErrClientSecretRequired = errors.New("auth0: client secret is required")
	ErrDomainRequired       = errors.New("auth0: domain is required")
)
