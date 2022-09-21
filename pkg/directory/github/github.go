// Package github contains a directory provider for github.
package github

// A ServiceAccount is used by the GitHub provider to query the GitHub API.
type ServiceAccount struct {
	Username            string
	PersonalAccessToken string
}

// see: https://docs.github.com/en/free-pro-team@latest/rest/reference/users#get-a-user
type apiUserObject struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
