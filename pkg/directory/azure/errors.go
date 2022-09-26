package azure

import "errors"

// Errors
var (
	ErrClientIDRequired     = errors.New("azure: client id is required")
	ErrClientSecretRequired = errors.New("azure: client secret is required")
	ErrDirectoryIDRequired  = errors.New("azure: directory id is required")
)
