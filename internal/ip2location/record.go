package ip2location

type (
	// A Record is a ip2location record.
	Record struct {
		Index    RecordIndex `json:"$index"`
		ID       string      `json:"id"`
		Country  string      `json:"country"`
		State    string      `json:"state"`
		City     string      `json:"city"`
		Zip      string      `json:"zip"`
		Timezone string      `json:"timezone"`
	}
	// A RecordIndex is how the record is indexed.
	RecordIndex struct {
		CIDR string `json:"cidr"`
	}
)
