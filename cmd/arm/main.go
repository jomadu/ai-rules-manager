package main

import (
	"fmt"
	"os"

	"github.com/max-dunn/ai-rules-manager/internal/cli"
	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// Build information injected by ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create root command with version info
	versionInfo := &cli.VersionInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
	rootCmd := cli.NewRootCommand(cfg, versionInfo)

	// Execute command
	return rootCmd.Execute()
}
