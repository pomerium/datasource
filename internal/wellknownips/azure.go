package wellknownips

import (
	"context"
	"encoding/json"
	"net/http"
)

// AzureIPRanges are the definitions of the ip ranges for services in Azure.
type AzureIPRanges struct {
	Values []struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		Properties struct {
			SystemService   string   `json:"systemService"`
			AddressPrefixes []string `json:"addressPrefixes"`
		} `json:"properties"`
	} `json:"values"`
}

// DefaultAzureIPRangesURL is the default Azure IP Ranges URL.
var DefaultAzureIPRangesURL = "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20220620.json"

// FetchAzureIPRanges fetches the Azure IP Ranges for all Azure services.
func FetchAzureIPRanges(
	ctx context.Context,
	client *http.Client,
	url string,
) (*AzureIPRanges, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ranges AzureIPRanges
	err = json.NewDecoder(res.Body).Decode(&ranges)
	if err != nil {
		return nil, err
	}

	return &ranges, nil
}

const (
	MicrosoftASNumber    = "8075"
	MicrosoftCountryCode = "US"
	MicrosoftASName      = "MICROSOFT-CORP-MSN-AS-BLOCK"
)

// RecordsFromAzureIPRanges converts AzureIPRanges into records.
func RecordsFromAzureIPRanges(in *AzureIPRanges) []Record {
	var records []Record
	for _, value := range in.Values {
		for _, prefix := range value.Properties.AddressPrefixes {
			records = append(records, Record{
				ID:          prefix,
				ASNumber:    MicrosoftASNumber,
				CountryCode: MicrosoftCountryCode,
				ASName:      MicrosoftASName,
				Service:     value.Name,
			})
		}
	}
	return records
}
