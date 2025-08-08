package main

import (
	"fmt"
	"os"

	"github.com/max-dunn/ai-rules-manager/internal/cli"
	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/version"
)

// Build information injected by ldflags
var (
	buildVersion   = "dev"
	buildCommit    = "unknown"
	buildTimestamp = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Set version information in version package
	version.Version = buildVersion
	version.Commit = buildCommit
	version.BuildTime = buildTimestamp

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create root command with version info
	versionInfo := &cli.VersionInfo{
		Version:   buildVersion,
		Commit:    buildCommit,
		BuildTime: buildTimestamp,
	}
	rootCmd := cli.NewRootCommand(cfg, versionInfo)

	// Execute command
	return rootCmd.Execute()
}
