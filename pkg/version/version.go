// Package version contains build-time version information.
package version

// These variables are set at build time using ldflags.
var (
	// Version is the semantic version of the addon.
	Version = "dev"
	// GitCommit is the git commit hash.
	GitCommit = "unknown"
	// BuildDate is the build date.
	BuildDate = "unknown"
)

// Info returns version information as a map.
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"gitCommit": GitCommit,
		"buildDate": BuildDate,
	}
}
