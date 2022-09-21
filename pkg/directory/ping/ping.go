// Package ping implements a directory provider for Ping.
package ping

// A ServiceAccount is used by the Ping provider to query the API.
type ServiceAccount struct {
	ClientID      string
	ClientSecret  string
	EnvironmentID string
}
