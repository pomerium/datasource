package client

type Host struct {
	ID         uint `json:"id"`
	HostIssues struct {
		FailingPoliciesCount         uint64  `json:"failing_policies_count"`
		CriticalVulnerabilitiesCount *uint64 `json:"critical_vulnerabilities_count,omitempty"`
		TotalIssuesCount             uint64  `json:"total_issues_count"`
	} `json:"issues,omitempty"`
}

type QueryReportItem struct {
	HostID  uint              `json:"host_id"`
	Columns map[string]string `json:"columns"`
}
