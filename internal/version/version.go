// Package version provides build-time version information for Rune.
package version

import "fmt"

// These are set at build time via -ldflags.
var (
	Version = "0.1.0"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns a human-readable version string.
func String() string {
	return fmt.Sprintf("Rune %s (%s, %s)", Version, Commit, Date)
}

// Short returns just the version number.
func Short() string {
	return fmt.Sprintf("Rune %s", Version)
}
