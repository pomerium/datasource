// Package azure contains an azure active directory directory provider.
package azure

// A ServiceAccount is used by the Azure provider to query the Microsoft Graph API.
type ServiceAccount struct {
	ClientID     string
	ClientSecret string
	DirectoryID  string
}
