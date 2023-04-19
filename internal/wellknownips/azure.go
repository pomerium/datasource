package wellknownips

import (
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/pomerium/datasource/internal/wellknownips/files"
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

// FetchAzureIPRanges fetches the Azure IP Ranges for all Azure services.
func FetchAzureIPRanges(
	_ context.Context,
) (*AzureIPRanges, error) {
	f, err := files.FS.Open("azure.json.gz")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var ranges AzureIPRanges
	err = json.NewDecoder(r).Decode(&ranges)
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
