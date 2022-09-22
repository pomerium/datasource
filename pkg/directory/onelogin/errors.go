package onelogin

import "errors"

// Errors
var (
	ErrClientIDRequired     = errors.New("onelogin: client id is required")
	ErrClientSecretRequired = errors.New("onelogin: client secret is requried")
)
