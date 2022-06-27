package wellknownips

import (
	"context"
	"encoding/json"
	"net/http"
)

// StripeIPRanges are the stripe IP ranges.
type StripeIPRanges struct {
	WebHooks []string `json:"WEBHOOKS"`
}

// The DefaultStripeIPRangesURL is the default stripe ip ranges url.
var DefaultStripeIPRangesURL = "https://stripe.com/files/ips/ips_webhooks.json"

// FetchStripeIPRanges fetches the stripe ip ranges.
func FetchStripeIPRanges(
	ctx context.Context,
	client *http.Client,
	url string,
) (*StripeIPRanges, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ranges StripeIPRanges
	err = json.NewDecoder(res.Body).Decode(&ranges)
	if err != nil {
		return nil, err
	}

	return &ranges, nil
}
