package client

import (
	"iter"
	"strconv"
)

type Host struct {
	ID                           string  `json:"id"`
	FailingPoliciesCount         uint64  `json:"failing_policies_count"`
	CriticalVulnerabilitiesCount *uint64 `json:"critical_vulnerabilities_count,omitempty"`
}

type hostRecord struct {
	ID         uint `json:"id"`
	HostIssues struct {
		FailingPoliciesCount         uint64  `json:"failing_policies_count"`
		CriticalVulnerabilitiesCount *uint64 `json:"critical_vulnerabilities_count,omitempty"`
		TotalIssuesCount             uint64  `json:"total_issues_count"`
	} `json:"issues,omitempty"`
}

func convertHostRecord(r hostRecord) (Host, error) {
	return Host{
		ID:                           strconv.FormatUint(uint64(r.ID), 10),
		FailingPoliciesCount:         r.HostIssues.FailingPoliciesCount,
		CriticalVulnerabilitiesCount: r.HostIssues.CriticalVulnerabilitiesCount,
	}, nil
}

type certificateQueryRecord struct {
	HostID  uint64 `json:"host_id"`
	Columns struct {
		SHA1 string `json:"sha1"`
	} `json:"columns"`
}

type CertificateSHA1QueryItem struct {
	HostID string `json:"host_id"`
	SHA1   string `json:"id"`
}

func convertCertificateQuery(c certificateQueryRecord) (CertificateSHA1QueryItem, error) {
	return CertificateSHA1QueryItem{
		HostID: strconv.FormatUint(c.HostID, 10),
		SHA1:   c.Columns.SHA1,
	}, nil
}

func convertIter2[T1, T2 any](
	iter1 iter.Seq2[T1, error],
	fn func(T1) (T2, error),
) iter.Seq2[T2, error] {
	return func(yield func(T2, error) bool) {
		for v1, err := range iter1 {
			var v2 T2
			if err != nil {
				if !yield(v2, err) {
					return
				}
				continue
			}
			v2, err := fn(v1)
			if err != nil {
				if !yield(v2, err) {
					return
				}
				continue
			}
			if !yield(v2, nil) {
				return
			}
		}
	}
}
