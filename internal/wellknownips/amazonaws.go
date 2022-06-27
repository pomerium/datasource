package wellknownips

import (
	"context"
	"encoding/json"
	"net/http"
)

type AmazonAWSIPRanges struct {
	Prefixes []struct {
		IPPrefix           string `json:"ip_prefix"`
		Region             string `json:"region"`
		Service            string `json:"service"`
		NetworkBorderGroup string `json:"network_border_group"`
	} `json:"prefixes"`
}

// DefaultAmazonAWSIPRangesURL is the default amazon aws ip ranges url.
var DefaultAmazonAWSIPRangesURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"

// FetchAmazonAWSIPRanges fetches the Amazon AWS IP Ranges.
func FetchAmazonAWSIPRanges(
	ctx context.Context,
	client *http.Client,
	url string,
) (*AmazonAWSIPRanges, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ranges AmazonAWSIPRanges
	err = json.NewDecoder(res.Body).Decode(&ranges)
	if err != nil {
		return nil, err
	}

	return &ranges, nil
}
