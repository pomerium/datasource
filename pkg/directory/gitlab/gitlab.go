// Package gitlab contains a directory provider for gitlab.
package gitlab

// A ServiceAccount is used by the Gitlab provider to query the Gitlab API.
type ServiceAccount struct {
	PrivateToken string
}
