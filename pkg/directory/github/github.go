// Package github contains a directory provider for github.
package github

// see: https://docs.github.com/en/free-pro-team@latest/rest/reference/users#get-a-user
type apiUserObject struct {
	NodeID string `json:"node_id"`
	Login  string `json:"login"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}
