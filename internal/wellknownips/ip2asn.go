package wellknownips

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type ip2asnRecord struct {
	RangeStart, RangeEnd string
	ASNumber             string
	CountryCode          string
	ASDescription        string
}

type ip2asnStream struct {
	rc io.ReadCloser

	init    sync.Once
	gr      *gzip.Reader
	scanner *bufio.Scanner

	record *ip2asnRecord
	err    error
}

// FetchIP2ASNDatabase fetches the IP2ASN database.
func FetchIP2ASNDatabase(
	ctx context.Context,
	client *http.Client,
	url string,
) (*ip2asnStream, error) { //nolint:revive
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ip2asn database: %w", err)
	}

	res, err := client.Do(req) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ip2asn database: %w", err)
	}

	return &ip2asnStream{
		rc: res.Body,
	}, nil
}

func (stream *ip2asnStream) Close() error {
	return stream.rc.Close()
}

func (stream *ip2asnStream) Next(ctx context.Context) bool {
	stream.init.Do(func() {
		stream.gr, stream.err = gzip.NewReader(stream.rc)
		if stream.err == nil {
			stream.scanner = bufio.NewScanner(stream.gr)
		}
	})
	if stream.err != nil {
		return false
	}

	if !stream.scanner.Scan() {
		stream.err = stream.scanner.Err()
		return false
	}

	parts := strings.Fields(stream.scanner.Text())
	// range_start range_end AS_number country_code AS_description
	if len(parts) < 5 {
		return false
	}

	stream.record = &ip2asnRecord{
		RangeStart:    parts[0],
		RangeEnd:      parts[1],
		ASNumber:      parts[2],
		CountryCode:   parts[3],
		ASDescription: parts[4],
	}
	return true
}

func (stream *ip2asnStream) Err() error {
	return stream.err
}

func (stream *ip2asnStream) Record() *ip2asnRecord {
	return stream.record
}
