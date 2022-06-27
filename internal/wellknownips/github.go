package wellknownips

import (
	"context"
	"encoding/json"
	"net/http"
)

// GitHubMeta is the GitHub metadata.
type GitHubMeta struct {
	Hooks      []string `json:"hooks"`
	Web        []string `json:"web"`
	API        []string `json:"api"`
	Git        []string `json:"git"`
	Packages   []string `json:"packages"`
	Pages      []string `json:"pages"`
	Importer   []string `json:"importer"`
	Actions    []string `json:"actions"`
	Dependabot []string `json:"dependabot"`
}

var DefaultGitHubMetaURL = "https://api.github.com/meta"

// FetchGitHubMeta fetches the GitHub metadata.
func FetchGitHubMeta(
	ctx context.Context,
	client *http.Client,
	url string,
) (*GitHubMeta, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var meta GitHubMeta
	err = json.NewDecoder(res.Body).Decode(&meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}
