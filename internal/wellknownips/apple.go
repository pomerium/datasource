package wellknownips

// AppleDomainVerificationIPAddresses are the domain verification ip addresses for Apple.
// https://developer.apple.com/documentation/apple_pay_on_the_web/setting_up_your_server#3179116
var AppleDomainVerificationIPAddresses = []string{
	"17.32.139.128/27",
	"17.32.139.160/27",
	"17.140.126.0/27",
	"17.140.126.32/27",
	"17.179.144.128/27",
	"17.179.144.160/27",
	"17.179.144.192/27",
	"17.179.144.224/27",
	"17.253.0.0/16",
}

const (
	// AppleASNumber is the Apple ASNumber.
	AppleASNumber    = "714"
	AppleCountryCode = "US"
	AppleASName      = "APPLE-ENGINEERING"
)

// RecordsFromAppleDomainVerificationIPAddresses converts AppleDomainVerificationIPAddresses into records.
func RecordsFromAppleDomainVerificationIPAddresses(in []string) []Record {
	var records []Record
	for _, prefix := range in {
		records = append(records, Record{
			ID:          prefix,
			ASNumber:    AppleASNumber,
			CountryCode: AppleCountryCode,
			ASName:      AppleASName,
		})
	}
	return records
}
