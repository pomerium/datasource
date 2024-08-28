package fleetdm

import (
	"context"
	"fmt"

	"github.com/pomerium/datasource/internal/fleetdm/client"
)

type Record struct {
	CertificateSHA1Fingerprint string `json:"id"`
	FailingPoliciesCount       int    `json:"failing_policies_count"`
}

func (srv *server) getRecords(
	ctx context.Context,
) ([]Record, error) {
	hosts, err := srv.client.ListHosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}

	certs, err := srv.client.QueryReport(ctx, srv.cfg.certificateQueryID)
	if err != nil {
		return nil, fmt.Errorf("query %d report: %w", srv.cfg.certificateQueryID, err)
	}

	hostIndex := make(map[uint]client.Host, len(hosts))
	for _, host := range hosts {
		hostIndex[host.ID] = host
	}

	records := make([]Record, 0, len(certs))
	for _, cert := range certs {
		host, ok := hostIndex[cert.HostID]
		if !ok {
			continue
		}

		records = append(records, Record{
			CertificateSHA1Fingerprint: cert.Columns["sha1"],
			FailingPoliciesCount:       int(host.HostIssues.FailingPoliciesCount),
		})
	}

	return records, nil
}
