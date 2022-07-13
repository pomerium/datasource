package wellknownips

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// AtlassianIPRanges are the Atlassian IP ranges.
type AtlassianIPRanges struct {
	Items []struct {
		CIDR    string   `json:"cidr"`
		Product []string `json:"product"`
	} `json:"items"`
}

// DefaultAtlassianIPRangesURL is the default Atlassian IP Ranges url.
var DefaultAtlassianIPRangesURL = "https://ip-ranges.atlassian.com/"

// FetchAtlassianIPRanges fetches the Atlassian IP Ranges.
func FetchAtlassianIPRanges(
	ctx context.Context,
	client *http.Client,
	url string,
) (*AtlassianIPRanges, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ranges AtlassianIPRanges
	err = json.NewDecoder(res.Body).Decode(&ranges)
	if err != nil {
		return nil, err
	}

	return &ranges, nil
}

const (
	AtlassianASNumber    = "133530"
	AtlassianCountryCode = "AU"
	AtlassianASName      = "ATLASSIAN PTY LTD"
)

// RecordsFromAtlassianIPRanges converts AtlassianIPRanges into records.
func RecordsFromAtlassianIPRanges(in *AtlassianIPRanges) []Record {
	var records []Record
	for _, item := range in.Items {
		records = append(records, Record{
			ID:          item.CIDR,
			ASNumber:    AtlassianASNumber,
			CountryCode: AtlassianCountryCode,
			ASName:      AtlassianASName,
			Service:     strings.Join(item.Product, " "),
		})
	}
	return records
}
