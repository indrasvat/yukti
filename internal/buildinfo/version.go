// Package buildinfo contains build-time information set by ldflags.
package buildinfo

// Build-time variables set via ldflags.
var (
	// Version is the semantic version (set by goreleaser or manually).
	Version = "dev"

	// Commit is the git commit hash.
	Commit = "unknown"

	// BuildDate is the build timestamp.
	BuildDate = "unknown"

	// GoVersion is the Go version used to build.
	GoVersion = "unknown"
)
