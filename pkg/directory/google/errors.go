package google

import "errors"

// Errors
var (
	ErrJSONKeyRequired         = errors.New("google: json key is required")
	ErrImpersonateUserRequired = errors.New("google: impersonate user is required")
)
