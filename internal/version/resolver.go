package version

// Build information - will be injected by ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// GetVersion returns the current ARM version
func GetVersion() string {
	return Version
}

// Version resolution and semantic versioning logic
// Will be implemented in task 5.1
