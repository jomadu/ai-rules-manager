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
	// Check if we need to initialize default files
	if shouldInitialize() {
		if err := initializeProject(); err != nil {
			return err
		}
		return nil
	}

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

// shouldInitialize checks if we should create default files
func shouldInitialize() bool {
	_, rulesExists := os.Stat("rules.json")
	_, armrcExists := os.Stat(".armrc")
	return os.IsNotExist(rulesExists) || os.IsNotExist(armrcExists)
}

// initializeProject creates default rules.json and .armrc files
func initializeProject() error {
	_, rulesExists := os.Stat("rules.json")
	_, armrcExists := os.Stat(".armrc")

	var created []string

	// Create rules.json if missing
	if os.IsNotExist(rulesExists) {
		manifest := &types.RulesManifest{
			Targets:      []string{},
			Dependencies: make(map[string]string),
		}
		if err := manifest.SaveManifestWithoutValidation("rules.json"); err != nil {
			return fmt.Errorf("failed to create rules.json: %w", err)
		}
		created = append(created, "rules.json")
	}

	// Create .armrc if missing
	if os.IsNotExist(armrcExists) {
		armrcContent := `# ARM Configuration File
# Configure registries for ruleset installation

# Example configurations (uncomment and modify as needed):

# GitLab Package Registry
# [sources]
# company = https://gitlab.company.com
# [sources.company]
# type = gitlab
# projectID = 12345
# authToken = $GITLAB_TOKEN

# AWS S3 Registry
# [sources]
# s3 = s3://my-rules-bucket/packages/
# [sources.s3]
# type = s3
# region = us-east-1
# authToken = $AWS_ACCESS_KEY_ID:$AWS_SECRET_ACCESS_KEY

# Git Repository
# [sources]
# awesome = https://github.com/user/awesome-rules
# [sources.awesome]
# type = git
# api = github
# authToken = $GITHUB_TOKEN

# HTTP Registry
# [sources]
# http = https://registry.example.com/
# [sources.http]
# type = http

# Default Public Registry
# [sources]
# default = https://registry.armjs.org/
`
		if err := os.WriteFile(".armrc", []byte(armrcContent), 0o644); err != nil {
			return fmt.Errorf("failed to create .armrc: %w", err)
		}
		created = append(created, ".armrc")
	}

	if len(created) > 0 {
		fmt.Printf("Created default configuration files: %s\n", strings.Join(created, ", "))
		fmt.Println("\nNext steps:")
		fmt.Println("1. Configure sources (GitLab registries, git repositories, etc.) in .armrc or use the arm config commands")
		fmt.Println("2. Set targets in rules.json (e.g., [\".cursorrules\", \".amazonq/rules\"])")
		fmt.Println("3. Add rulesets to rules.json dependencies or use the arm install commands")
	}

	return nil
}
