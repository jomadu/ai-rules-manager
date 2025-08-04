package main

import (
	"fmt"
	"strings"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/performance"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [ruleset]",
	Short: "Install a ruleset or install from manifest",
	Long:  "Install a specific ruleset or install all dependencies from rules.json manifest",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return installFromManifest()
	}
	return installRuleset(args[0])
}

func installFromManifest() error {
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Load configuration and create registry manager
	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	registryManager := registry.NewManager(configManager)

	// Prepare download jobs
	var jobs []performance.DownloadJob
	for name, versionSpec := range manifest.Dependencies {
		registryName := registryManager.ParseRegistryName(name)
		cleanName := registryManager.StripRegistryPrefix(name)

		jobs = append(jobs, performance.DownloadJob{
			Name:            name,
			VersionSpec:     versionSpec,
			RegistryName:    registryName,
			CleanName:       cleanName,
			RegistryManager: registryManager,
		})
	}

	// Download in parallel
	downloader := performance.NewParallelDownloader(registryManager)
	results := downloader.DownloadAll(jobs)

	// Print results and return error if any failed
	return performance.PrintResults(results)
}

func installRuleset(rulesetSpec string) error {
	name, version := parseRulesetSpec(rulesetSpec)
	fmt.Printf("Installing %s@%s...\n", name, version)

	// Load configuration and create registry manager
	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	registryManager := registry.NewManager(configManager)

	// Get registry name and clean name
	registryName := registryManager.ParseRegistryName(name)
	cleanName := registryManager.StripRegistryPrefix(name)

	// Create installer with caching support
	installer := installer.NewWithManager(registryManager, registryName, cleanName)
	return installer.Install(cleanName, version)
}

// parseRulesetSpec parses "name@version" or just "name" (defaults to latest)
func parseRulesetSpec(spec string) (name, version string) {
	parts := strings.Split(spec, "@")
	if len(parts) == 1 {
		return parts[0], "latest"
	}
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// Handle org@package@version format
	if len(parts) == 3 {
		return fmt.Sprintf("%s@%s", parts[0], parts[1]), parts[2]
	}
	return spec, "latest"
}
