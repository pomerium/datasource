package ping

import "errors"

// Errors
var (
	ErrEnvironmentIDRequired = errors.New("ping: environment id is required")
	ErrClientIDRequired      = errors.New("ping: client id is required")
	ErrClientSecretRequired  = errors.New("ping: client secret is requried")
)
