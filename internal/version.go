package internal

import (
	"strings"
)

var (
	// ProjectName is the canonical project name set by ldl flags
	ProjectName = "pomerium-ext-data-agent"
	// ProjectURL is the canonical project url  set by ldl flags
	ProjectURL = ""
	// Version specifies Semantic versioning increment (MAJOR.MINOR.PATCH).
	Version = "v0.0.0"
	// GitCommit specifies the git commit sha, set by the compiler.
	GitCommit = ""
	// BuildMeta specifies release type (dev,rc1,beta,etc)
	BuildMeta = ""
)

// FullVersion returns a version string.
func FullVersion() string {
	var sb strings.Builder
	sb.Grow(len(Version) + len(GitCommit) + len(BuildMeta) + len("-") + len("+"))
	sb.WriteString(Version)
	if BuildMeta != "" {
		sb.WriteString("-" + BuildMeta)
	}
	if GitCommit != "" {
		sb.WriteString("+" + GitCommit)
	}
	return sb.String()
}