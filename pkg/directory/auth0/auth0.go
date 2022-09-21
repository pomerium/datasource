package auth0

// A ServiceAccount is used by the Auth0 provider to query the API.
type ServiceAccount struct {
	Domain       string
	ClientID     string
	ClientSecret string
}
