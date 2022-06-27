package wellknownips

import (
	"encoding/json"
	"net/netip"

	"github.com/pomerium/datasource/internal/netutil"
)

// A Record is a Well-Known IP Record.
type Record struct {
	ID          string
	ASNumber    string
	CountryCode string
	ASName      string
	Service     string
}

// RecordsFromIP2ASNRecord converts IP2ASN records to Well-Known IP Records.
func RecordsFromIP2ASNRecord(in *ip2asnRecord) []Record {
	if in.ASNumber == "0" {
		return nil
	}

	rangeStart, err := netip.ParseAddr(in.RangeStart)
	if err != nil {
		return nil
	}

	rangeEnd, err := netip.ParseAddr(in.RangeEnd)
	if err != nil {
		return nil
	}

	var records []Record
	prefixes := netutil.AddrRangeToPrefixes(rangeStart, rangeEnd)
	for _, prefix := range prefixes {
		records = append(records, Record{
			ID:          prefix.String(),
			ASNumber:    in.ASNumber,
			CountryCode: in.CountryCode,
			ASName:      in.ASDescription,
		})
	}
	return records
}

// MarshalJSON marshals the Well-Known IP Record as a JSON object.
func (record Record) MarshalJSON() ([]byte, error) {
	var x struct {
		Index struct {
			CIDR string `json:"cidr"`
		} `json:"index"`
		ID          string `json:"id"`
		ASNumber    string `json:"as_number"`
		CountryCode string `json:"country_code"`
		ASName      string `json:"as_name"`
		Service     string `json:"service,omitempty"`
	}
	x.Index.CIDR = record.ID
	x.ID = record.ID
	x.ASNumber = record.ASNumber
	x.CountryCode = record.CountryCode
	x.ASName = record.ASName
	x.Service = record.Service
	return json.Marshal(x)
}
