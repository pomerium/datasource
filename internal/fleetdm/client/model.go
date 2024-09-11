package client

import (
	"iter"
	"maps"
	"slices"
	"strconv"
	"time"
)

type Host struct {
	ID                   string    `json:"id"`
	Seen                 time.Time `json:"seen_time"`
	FailingPoliciesCount uint64    `json:"failing_policies_count"`
	// FailingCriticalPoliciesCount is the number of critical policies that the host is failing. This is a calculated value.
	FailingCriticalPoliciesCount uint64  `json:"failing_critical_policies_count"`
	CriticalVulnerabilitiesCount *uint64 `json:"critical_vulnerabilities_count,omitempty"`
	PoliciesPassing              []uint  `json:"policies_passing,omitempty"`
	PoliciesFailing              []uint  `json:"policies_failing,omitempty"`
	// CVEs is a map of CVE to whether the host is vulnerable to the CVE
	CVEs []string `json:"cves,omitempty"`
}

type HostPolicyStatus struct {
	ID       uint   `json:"id"`
	Response string `json:"response"`
}

type hostRecord struct {
	ID         uint      `json:"id"`
	Seen       time.Time `json:"seen_time"`
	HostIssues struct {
		FailingPoliciesCount         uint64  `json:"failing_policies_count"`
		CriticalVulnerabilitiesCount *uint64 `json:"critical_vulnerabilities_count,omitempty"`
		TotalIssuesCount             uint64  `json:"total_issues_count"`
	} `json:"issues,omitempty"`
	Policies []struct {
		ID       uint   `json:"id"`
		Response string `json:"response"`
		Critical bool   `json:"critical"`
	} `json:"policies"`
	Software []struct {
		Vulnerabilities []struct {
			CVE string `json:"cve"`
		} `json:"vulnerabilities"`
	} `json:"software"`
}

func convertHostRecord(r hostRecord) (Host, error) {
	var policiesPassing, policiesFailing []uint
	failingCriticalPoliciesCount := uint64(0)

	for _, p := range r.Policies {
		passing := p.Response == "pass"
		failing := p.Response == "fail"

		if p.Critical && failing {
			failingCriticalPoliciesCount++
		}

		if passing {
			policiesPassing = append(policiesPassing, p.ID)
		}
		if failing {
			policiesFailing = append(policiesFailing, p.ID)
		}
	}

	cves := make(map[string]bool)
	for _, s := range r.Software {
		for _, v := range s.Vulnerabilities {
			cves[v.CVE] = true
		}
	}

	return Host{
		ID:                           strconv.FormatUint(uint64(r.ID), 10),
		FailingPoliciesCount:         r.HostIssues.FailingPoliciesCount,
		FailingCriticalPoliciesCount: failingCriticalPoliciesCount,
		CriticalVulnerabilitiesCount: r.HostIssues.CriticalVulnerabilitiesCount,
		PoliciesPassing:              policiesPassing,
		PoliciesFailing:              policiesFailing,
		CVEs:                         slices.Collect(maps.Keys(cves)),
		Seen:                         r.Seen,
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

type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resolution  string `json:"resolution"`
}

func (p Policy) GetID() string {
	return p.ID
}

type policyRecord struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resolution  string `json:"resolution"`
}

func convertPolicy(r policyRecord) (Policy, error) {
	return Policy{
		ID:          strconv.FormatUint(uint64(r.ID), 10),
		Name:        r.Name,
		Description: r.Description,
		Resolution:  r.Resolution,
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
