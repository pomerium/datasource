package github

import "errors"

// Errors
var (
	ErrPersonalAccessTokenRequired = errors.New("github: personal access token is required")
	ErrUsernameRequired            = errors.New("github: username is required")
)
