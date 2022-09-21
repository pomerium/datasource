// Package onelogin contains the onelogin directory provider.
package onelogin

// A ServiceAccount is used by the OneLogin provider to query the API.
type ServiceAccount struct {
	ClientID     string
	ClientSecret string
}
