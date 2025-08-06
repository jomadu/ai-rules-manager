package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/performance"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [ruleset[@version]]",
	Short: "Install rulesets from registries",
	Long: `Install rulesets from configured registries or from rules.json manifest.

When no arguments are provided, installs all dependencies from rules.json.
When a ruleset is specified, installs that specific ruleset.

Examples:
  arm install                     # Install from rules.json manifest
  arm install typescript-rules    # Install latest version
  arm install company@rules@1.0   # Install specific version from registry
  arm install @company/rules       # Install from company registry`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInstall,
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
	// Check configuration state
	hasManifest := fileExists("rules.json")
	hasConfig := hasValidConfig()

	// Scenario 1a: No configuration files
	if !hasManifest && !hasConfig {
		if err := createStubFiles(); err != nil {
			return err
		}
		fmt.Println("Created stub configuration files (.armrc and rules.json).")
		fmt.Println("Please configure your registries in .armrc and add dependencies to rules.json.")
		return nil
	}

	// Scenario 3a: Has .armrc but no rules.json
	if !hasManifest && hasConfig {
		if err := createStubManifest(); err != nil {
			return err
		}
		fmt.Println("Created stub rules.json file.")
		fmt.Println("Please add dependencies to rules.json and run 'arm install' again.")
		return nil
	}

	// Scenario 4a: Has rules.json but no .armrc
	if hasManifest && !hasConfig {
		if err := createStubConfig(); err != nil {
			return err
		}
		cwd, _ := os.Getwd()
		return fmt.Errorf("No registry sources configured. Please configure a source in .armrc file.\nCreated stub .armrc file in %s for you to customize.", cwd)
	}

	// Scenario 2a: Both files exist - normal installation
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	registryManager := registry.NewManager(configManager)

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

	downloader := performance.NewParallelDownloader(registryManager)
	results := downloader.DownloadAll(jobs)

	return performance.PrintResults(results)
}

func installRuleset(rulesetSpec string) error {
	name, version := parseRulesetSpec(rulesetSpec)

	// Check configuration state
	hasManifest := fileExists("rules.json")
	hasConfig := hasValidConfig()

	// Scenario 1b: No configuration files
	if !hasManifest && !hasConfig {
		if err := createStubFiles(); err != nil {
			return err
		}
		fmt.Println("Created stub configuration files (.armrc and rules.json).")
		fmt.Printf("Ruleset %s@%s was not installed due to missing source configuration.\n", name, version)
		fmt.Println("Please configure your registries in .armrc and run the install command again.")
		return nil
	}

	// Scenario 3b: Has .armrc but no rules.json
	if !hasManifest && hasConfig {
		if err := createManifestWithRuleset(name, version); err != nil {
			return err
		}
		fmt.Println("Created rules.json file with the specified ruleset.")
		// Continue to install the ruleset
	}

	// Scenario 4b: Has rules.json but no .armrc
	if hasManifest && !hasConfig {
		if err := createStubConfig(); err != nil {
			return err
		}
		cwd, _ := os.Getwd()
		return fmt.Errorf("No registry sources configured. Please configure a source in .armrc file.\nCreated stub .armrc file in %s for you to customize.", cwd)
	}

	// Scenario 2b or 3b (after creating manifest): Both files exist - normal installation
	fmt.Printf("Installing %s@%s...\n", name, version)

	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	registryManager := registry.NewManager(configManager)
	registryName := registryManager.ParseRegistryName(name)
	cleanName := registryManager.StripRegistryPrefix(name)

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

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// hasValidConfig checks if there's a valid .armrc configuration
func hasValidConfig() bool {
	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return false
	}
	config := configManager.GetConfig()
	return config != nil && len(config.Sources) > 0
}

// createStubConfig creates a stub .armrc file
func createStubConfig() error {
	stubContent := `# Example configuration for ARM (AI Rules Manager)
# Uncomment and modify the sections below to configure your registries

# [sources]
# my-rules = https://github.com/username/my-rules

# [sources.my-rules]
# type = git
# api = github
# # authToken = $GITHUB_TOKEN  # Optional for private repos
`
	return os.WriteFile(".armrc", []byte(stubContent), 0o644)
}

// createStubManifest creates a stub rules.json file
func createStubManifest() error {
	manifest := &types.RulesManifest{
		Targets:      types.GetDefaultTargets(),
		Dependencies: make(map[string]string),
	}
	return manifest.SaveManifest("rules.json")
}

// createManifestWithRuleset creates rules.json with the specified ruleset
func createManifestWithRuleset(name, version string) error {
	manifest := &types.RulesManifest{
		Targets:      types.GetDefaultTargets(),
		Dependencies: map[string]string{name: version},
	}
	return manifest.SaveManifest("rules.json")
}

// createStubFiles creates both .armrc and rules.json stub files
func createStubFiles() error {
	if err := createStubConfig(); err != nil {
		return fmt.Errorf("failed to create .armrc: %w", err)
	}
	if err := createStubManifest(); err != nil {
		return fmt.Errorf("failed to create rules.json: %w", err)
	}
	return nil
}
