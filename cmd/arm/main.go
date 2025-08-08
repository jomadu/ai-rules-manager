package main

import (
	"fmt"
	"os"

	"github.com/max-dunn/ai-rules-manager/internal/cli"
	"github.com/max-dunn/ai-rules-manager/internal/config"
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

	// Create root command
	rootCmd := cli.NewRootCommand(cfg)

	// Execute command
	return rootCmd.Execute()
}
