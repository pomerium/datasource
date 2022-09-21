// Package google contains the Google directory provider.
package google

// A ServiceAccount is used to authenticate with the Google APIs.
//
// Google oauth fields are from https://github.com/golang/oauth2/blob/master/google/google.go#L99
type ServiceAccount struct {
	Type string `json:"type"` // serviceAccountKey or userCredentialsKey

	// Service Account fields
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`

	// User Credential fields
	// (These typically come from gcloud auth.)
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`

	// The User to use for Admin Directory API calls
	ImpersonateUser string `json:"impersonate_user"`
}
