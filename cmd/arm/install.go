package main

import (
	"fmt"
	"strings"

	"github.com/jomadu/arm/internal/installer"
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

	// TODO: Make registry URL configurable
	registryURL := "http://localhost:8080"
	installer := installer.New(registryURL)

	for name, versionSpec := range manifest.Dependencies {
		fmt.Printf("Installing %s@%s...\n", name, versionSpec)
		if err := installer.Install(name, versionSpec); err != nil {
			return fmt.Errorf("failed to install %s: %w", name, err)
		}
	}

	fmt.Println("All dependencies installed successfully")
	return nil
}

func installRuleset(rulesetSpec string) error {
	name, version := parseRulesetSpec(rulesetSpec)
	fmt.Printf("Installing %s@%s...\n", name, version)

	// TODO: Make registry URL configurable
	registryURL := "http://localhost:8080"
	installer := installer.New(registryURL)

	return installer.Install(name, version)
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
