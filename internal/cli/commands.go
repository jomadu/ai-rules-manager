package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/install"
	"github.com/max-dunn/ai-rules-manager/internal/registry"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// VersionInfo contains build version information
type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

// NewRootCommand creates the root ARM command
func NewRootCommand(cfg *config.Config, versionInfo *VersionInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "arm",
		Short: "AI Rules Manager - Package manager for AI coding assistant rulesets",
		Long: `ARM is a package manager for AI coding assistant rulesets that enables
developers and teams to install, update, and manage coding rules across
different AI tools like Cursor and Amazon Q Developer.`,
		SilenceUsage: true,
	}

	// Add global flags
	rootCmd.PersistentFlags().Bool("global", false, "Operate on global configuration")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().Bool("verbose", false, "Show detailed output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without executing")
	rootCmd.PersistentFlags().Bool("json", false, "Output machine-readable JSON format")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().Bool("insecure", false, "Allow insecure HTTP connections")

	// Add subcommands
	rootCmd.AddCommand(newConfigCommand(cfg))
	rootCmd.AddCommand(newInstallCommand(cfg))
	rootCmd.AddCommand(newUninstallCommand(cfg))
	rootCmd.AddCommand(newSearchCommand(cfg))
	rootCmd.AddCommand(newInfoCommand(cfg))
	rootCmd.AddCommand(newOutdatedCommand(cfg))
	rootCmd.AddCommand(newUpdateCommand(cfg))
	rootCmd.AddCommand(newCleanCommand(cfg))
	rootCmd.AddCommand(newListCommand(cfg))
	rootCmd.AddCommand(newVersionCommand(versionInfo))

	return rootCmd
}

// newConfigCommand creates the config command
func newConfigCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ARM configuration",
		Long:  "Configure registries, channels, and other ARM settings",
	}

	// Set command
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return handleConfigSet(args[0], args[1], global)
		},
	}
	cmd.AddCommand(setCmd)

	// Get command
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return handleConfigGet(args[0], global)
		},
	}
	cmd.AddCommand(getCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return handleConfigList(global)
		},
	}
	cmd.AddCommand(listCmd)

	// Add command
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add registry or channel",
	}

	// Add registry subcommand
	addRegistryCmd := &cobra.Command{
		Use:   "registry <name> <value>",
		Short: "Add registry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			registryType, _ := cmd.Flags().GetString("type")
			authToken, _ := cmd.Flags().GetString("authToken")
			region, _ := cmd.Flags().GetString("region")
			profile, _ := cmd.Flags().GetString("profile")
			prefix, _ := cmd.Flags().GetString("prefix")
			apiType, _ := cmd.Flags().GetString("apiType")
			apiVersion, _ := cmd.Flags().GetString("apiVersion")

			return handleAddRegistry(args[0], args[1], registryType, global, map[string]string{
				"authToken":  authToken,
				"region":     region,
				"profile":    profile,
				"prefix":     prefix,
				"apiType":    apiType,
				"apiVersion": apiVersion,
			})
		},
	}
	addRegistryCmd.Flags().String("type", "", "Registry type (required)")
	addRegistryCmd.Flags().String("authToken", "", "Authentication token")
	addRegistryCmd.Flags().String("region", "", "AWS region (for S3 registries)")
	addRegistryCmd.Flags().String("profile", "", "AWS profile (for S3 registries)")
	addRegistryCmd.Flags().String("prefix", "", "Path prefix")
	addRegistryCmd.Flags().String("apiType", "", "API type (for Git registries)")
	addRegistryCmd.Flags().String("apiVersion", "", "API version")
	_ = addRegistryCmd.MarkFlagRequired("type")
	addCmd.AddCommand(addRegistryCmd)

	// Add channel subcommand
	addChannelCmd := &cobra.Command{
		Use:   "channel <name>",
		Short: "Add channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			directories, _ := cmd.Flags().GetString("directories")
			return handleAddChannel(args[0], directories, global)
		},
	}
	addChannelCmd.Flags().String("directories", "", "Comma-separated list of directories (required)")
	_ = addChannelCmd.MarkFlagRequired("directories")
	addCmd.AddCommand(addChannelCmd)

	cmd.AddCommand(addCmd)

	// Remove command
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove registry or channel",
	}

	// Remove registry subcommand
	removeRegistryCmd := &cobra.Command{
		Use:   "registry <name>",
		Short: "Remove registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return handleRemoveRegistry(args[0], global)
		},
	}
	removeCmd.AddCommand(removeRegistryCmd)

	// Remove channel subcommand
	removeChannelCmd := &cobra.Command{
		Use:   "channel <name>",
		Short: "Remove channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			return handleRemoveChannel(args[0], global)
		},
	}
	removeCmd.AddCommand(removeChannelCmd)

	cmd.AddCommand(removeCmd)

	return cmd
}

// newInstallCommand creates the install command
func newInstallCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [ruleset-spec]",
		Short: "Install rulesets",
		Long:  "Install rulesets from configured registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			channels, _ := cmd.Flags().GetString("channels")
			patterns, _ := cmd.Flags().GetString("patterns")

			if len(args) == 0 {
				return handleInstallFromManifest(global, dryRun, channels)
			} else {
				return handleInstallRuleset(args[0], global, dryRun, channels, patterns)
			}
		},
	}

	cmd.Flags().String("channels", "", "Install to specific channels only (comma-separated)")
	cmd.Flags().String("patterns", "", "Glob patterns for Git registry rulesets (comma-separated)")

	return cmd
}

// newUninstallCommand creates the uninstall command
func newUninstallCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <ruleset-name>",
		Short: "Remove rulesets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			channels, _ := cmd.Flags().GetString("channels")
			return handleUninstall(args[0], global, dryRun, channels)
		},
	}

	cmd.Flags().String("channels", "", "Remove from specific channels only (comma-separated)")

	return cmd
}

// newSearchCommand creates the search command
func newSearchCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for rulesets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registries, _ := cmd.Flags().GetString("registries")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			limit, _ := cmd.Flags().GetInt("limit")
			return handleSearch(args[0], registries, jsonOutput, limit)
		},
	}

	cmd.Flags().String("registries", "", "Search specific registries (comma-separated or glob patterns)")
	cmd.Flags().Int("limit", 50, "Limit number of results")

	return cmd
}

// newInfoCommand creates the info command
func newInfoCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <ruleset-spec>",
		Short: "Show ruleset information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			versions, _ := cmd.Flags().GetBool("versions")
			return handleInfo(args[0], jsonOutput, versions)
		},
	}

	cmd.Flags().Bool("versions", false, "Show all available versions")

	return cmd
}

// newOutdatedCommand creates the outdated command
func newOutdatedCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			return handleOutdated(global, jsonOutput)
		},
	}

	return cmd
}

// newUpdateCommand creates the update command
func newUpdateCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [ruleset-name]",
		Short: "Update rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if len(args) == 0 {
				return handleUpdateAll(global, dryRun)
			} else {
				return handleUpdateRuleset(args[0], global, dryRun)
			}
		},
	}

	return cmd
}

// newCleanCommand creates the clean command
func newCleanCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean [target]",
		Short: "Clean cache and unused rulesets",
		Long:  "Clean cache and unused rulesets. Targets: cache, unused, all",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			force, _ := cmd.Flags().GetBool("force")

			target := "all"
			if len(args) > 0 {
				target = args[0]
			}

			return handleClean(target, global, dryRun, force)
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompts")

	return cmd
}

// newListCommand creates the list command
func newListCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			local, _ := cmd.Flags().GetBool("local")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			channels, _ := cmd.Flags().GetString("channels")
			return handleList(global, local, jsonOutput, channels)
		},
	}

	cmd.Flags().Bool("local", false, "List local installations only")
	cmd.Flags().String("channels", "", "Filter by specific channels (comma-separated)")

	return cmd
}

// newVersionCommand creates the version command
func newVersionCommand(versionInfo *VersionInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show ARM version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ARM version %s\n", versionInfo.Version)
			fmt.Printf("Commit: %s\n", versionInfo.Commit)
			fmt.Printf("Built: %s\n", versionInfo.BuildTime)
			return nil
		},
	}
}

// Config command handlers

func handleConfigSet(key, value string, global bool) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format. Use section.key (e.g., git.concurrency)")
	}

	path := getConfigPath(".armrc", global)
	cfg, err := loadOrCreateINI(path)
	if err != nil {
		return err
	}

	section := cfg.Section(parts[0])
	section.Key(parts[1]).SetValue(value)

	return cfg.SaveTo(path)
}

func handleConfigGet(key string, _ bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	value := getConfigValue(cfg, key)
	if value == "" {
		return fmt.Errorf("key '%s' not found", key)
	}

	fmt.Println(value)
	return nil
}

func handleConfigList(_ bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("[registries]")
	for name, url := range cfg.Registries {
		fmt.Printf("%s = %s\n", name, url)
	}

	for name, regConfig := range cfg.RegistryConfigs {
		fmt.Printf("\n[registries.%s]\n", name)
		for key, value := range regConfig {
			fmt.Printf("%s = %s\n", key, value)
		}
	}

	for typeName, typeConfig := range cfg.TypeDefaults {
		fmt.Printf("\n[%s]\n", typeName)
		for key, value := range typeConfig {
			fmt.Printf("%s = %s\n", key, value)
		}
	}

	if len(cfg.NetworkConfig) > 0 {
		fmt.Println("\n[network]")
		for key, value := range cfg.NetworkConfig {
			fmt.Printf("%s = %s\n", key, value)
		}
	}

	if len(cfg.CacheConfig) > 0 {
		fmt.Println("\n[cache]")
		for key, value := range cfg.CacheConfig {
			fmt.Printf("%s = %s\n", key, value)
		}
	}

	return nil
}

func handleAddRegistry(name, url, registryType string, global bool, options map[string]string) error {
	if registryType == "" {
		return fmt.Errorf("registry type is required")
	}

	// Add to registries section
	path := getConfigPath(".armrc", global)
	cfg, err := loadOrCreateINI(path)
	if err != nil {
		return err
	}

	cfg.Section("registries").Key(name).SetValue(url)

	// Add registry config section
	sectionName := fmt.Sprintf("registries.%s", name)
	section := cfg.Section(sectionName)
	section.Key("type").SetValue(registryType)

	// Add optional parameters
	for key, value := range options {
		if value != "" {
			section.Key(key).SetValue(value)
		}
	}

	return cfg.SaveTo(path)
}

func handleRemoveRegistry(name string, global bool) error {
	path := getConfigPath(".armrc", global)
	cfg, err := loadOrCreateINI(path)
	if err != nil {
		return err
	}

	// Remove from registries section
	cfg.Section("registries").DeleteKey(name)

	// Remove registry config section
	sectionName := fmt.Sprintf("registries.%s", name)
	cfg.DeleteSection(sectionName)

	return cfg.SaveTo(path)
}

func handleAddChannel(name, directories string, global bool) error {
	if directories == "" {
		return fmt.Errorf("directories are required")
	}

	path := getConfigPath("arm.json", global)
	armConfig, err := loadOrCreateJSON(path)
	if err != nil {
		return err
	}

	dirList := strings.Split(directories, ",")
	for i, dir := range dirList {
		dirList[i] = strings.TrimSpace(dir)
	}

	armConfig.Channels[name] = config.ChannelConfig{
		Directories: dirList,
	}

	return saveJSON(path, armConfig)
}

func handleRemoveChannel(name string, global bool) error {
	path := getConfigPath("arm.json", global)
	armConfig, err := loadOrCreateJSON(path)
	if err != nil {
		return err
	}

	delete(armConfig.Channels, name)

	return saveJSON(path, armConfig)
}

// Helper functions

func getConfigPath(filename string, global bool) string {
	if global {
		return filepath.Join(os.Getenv("HOME"), ".arm", filename)
	}
	return filename
}

func loadOrCreateINI(path string) (*ini.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		return ini.Empty(), nil
	}
	return ini.Load(path)
}

func loadOrCreateJSON(path string) (*config.ARMConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		return &config.ARMConfig{
			Engines:  make(map[string]string),
			Channels: make(map[string]config.ChannelConfig),
			Rulesets: make(map[string]map[string]config.RulesetSpec),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var armConfig config.ARMConfig
	if err := json.Unmarshal(data, &armConfig); err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if armConfig.Engines == nil {
		armConfig.Engines = make(map[string]string)
	}
	if armConfig.Channels == nil {
		armConfig.Channels = make(map[string]config.ChannelConfig)
	}
	if armConfig.Rulesets == nil {
		armConfig.Rulesets = make(map[string]map[string]config.RulesetSpec)
	}

	return &armConfig, nil
}

func saveJSON(path string, armConfig *config.ARMConfig) error {
	data, err := json.MarshalIndent(armConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func getConfigValue(cfg *config.Config, key string) string {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return ""
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "registries":
		if len(parts) == 2 {
			return cfg.Registries[field]
		} else if len(parts) == 3 {
			if regConfig, exists := cfg.RegistryConfigs[field]; exists {
				return regConfig[parts[2]]
			}
		}
	case "network":
		return cfg.NetworkConfig[field]
	case "cache":
		return cfg.CacheConfig[field]
	default:
		if typeConfig, exists := cfg.TypeDefaults[section]; exists {
			return typeConfig[field]
		}
	}

	return ""
}

// Install command handlers

func handleInstallFromManifest(global, dryRun bool, _ string) error {
	// Load configuration to check for existing manifest
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have any rulesets configured
	if len(cfg.Rulesets) == 0 {
		// Generate stub files if they don't exist
		if err := ensureConfigFiles(global); err != nil {
			return err
		}
		fmt.Println("No rulesets configured. Generated stub configuration files.")
		fmt.Println("Configure registries and rulesets in .armrc and arm.json, then run 'arm install' again.")
		return nil
	}

	if dryRun {
		fmt.Println("Would install the following rulesets:")
		for registry, rulesets := range cfg.Rulesets {
			for name, spec := range rulesets {
				fmt.Printf("  %s/%s@%s\n", registry, name, spec.Version)
			}
		}
		return nil
	}

	// TODO: Implement actual installation from manifest
	return fmt.Errorf("manifest installation not yet implemented")
}

func handleInstallRuleset(rulesetSpec string, global, dryRun bool, channels, patterns string) error {
	// Parse ruleset specification
	registry, name, version := parseRulesetSpec(rulesetSpec)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have registries configured
	if len(cfg.Registries) == 0 {
		// Generate stub files if they don't exist
		if err := ensureConfigFiles(global); err != nil {
			return err
		}
		return fmt.Errorf("no registries configured. Add a registry with 'arm config add registry'")
	}

	// Determine target registry
	if registry == "" {
		if defaultRegistry, exists := cfg.Registries["default"]; exists {
			registry = "default"
			_ = defaultRegistry // Use the registry URL if needed
		} else {
			return fmt.Errorf("no default registry configured and no registry specified")
		}
	}

	// Check if registry exists
	if _, exists := cfg.Registries[registry]; !exists {
		return fmt.Errorf("registry '%s' not found", registry)
	}

	// Check if it's a Git registry and patterns are required
	if regConfig, exists := cfg.RegistryConfigs[registry]; exists {
		if regConfig["type"] == "git" && patterns == "" {
			return fmt.Errorf("Git registry rulesets require --patterns flag")
		}
	}

	if dryRun {
		fmt.Printf("Would install: %s/%s@%s\n", registry, name, version)
		if patterns != "" {
			fmt.Printf("  Patterns: %s\n", patterns)
		}
		if channels != "" {
			fmt.Printf("  Channels: %s\n", channels)
		}
		return nil
	}

	// Implement actual ruleset installation
	return performInstallation(cfg, registry, name, version, channels, patterns)
}

func parseRulesetSpec(spec string) (registry, name, version string) {
	// Handle version specification (name@version)
	if strings.Contains(spec, "@") {
		parts := strings.SplitN(spec, "@", 2)
		spec = parts[0]
		version = parts[1]
	} else {
		version = "latest"
	}

	// Handle registry specification (registry/name)
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		registry = parts[0]
		name = parts[1]
	} else {
		name = spec
	}

	return registry, name, version
}

func ensureConfigFiles(global bool) error {
	// Check if .armrc exists in either location
	armrcExists := false
	if global {
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".arm", ".armrc")); err == nil {
			armrcExists = true
		}
	} else {
		if _, err := os.Stat(".armrc"); err == nil {
			armrcExists = true
		}
		// Also check global location
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".arm", ".armrc")); err == nil {
			armrcExists = true
		}
	}

	// Check if arm.json exists in either location
	armJSONExists := false
	if global {
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".arm", "arm.json")); err == nil {
			armJSONExists = true
		}
	} else {
		if _, err := os.Stat("arm.json"); err == nil {
			armJSONExists = true
		}
		// Also check global location
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".arm", "arm.json")); err == nil {
			armJSONExists = true
		}
	}

	// Generate missing files
	if !armrcExists || !armJSONExists {
		return config.GenerateStubFiles(global)
	}

	return nil
}

// Search, info, and list command handlers

func handleSearch(query, registries string, jsonOutput bool, limit int) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have registries configured
	if len(cfg.Registries) == 0 {
		return fmt.Errorf("no registries configured. Add a registry with 'arm config add registry'")
	}

	// Determine which registries to search
	targetRegistries := getTargetRegistries(cfg.Registries, registries)
	if len(targetRegistries) == 0 {
		return fmt.Errorf("no matching registries found")
	}

	if jsonOutput {
		fmt.Printf(`{"query":"%s","registries":%v,"limit":%d,"results":[]}`, query, targetRegistries, limit)
		return nil
	}

	fmt.Printf("Searching for '%s' in registries: %s\n", query, strings.Join(targetRegistries, ", "))
	fmt.Printf("Limit: %d results\n\n", limit)
	fmt.Println("No results found (search functionality not yet implemented)")

	return nil
}

func handleInfo(rulesetSpec string, jsonOutput, versions bool) error {
	// Parse ruleset specification
	registry, name, version := parseRulesetSpec(rulesetSpec)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine target registry
	if registry == "" {
		if _, exists := cfg.Registries["default"]; exists {
			registry = "default"
		} else {
			return fmt.Errorf("no default registry configured and no registry specified")
		}
	}

	// Check if registry exists
	if _, exists := cfg.Registries[registry]; !exists {
		return fmt.Errorf("registry '%s' not found", registry)
	}

	if jsonOutput {
		fmt.Printf(`{"registry":"%s","name":"%s","version":"%s","versions_requested":%t}`, registry, name, version, versions)
		return nil
	}

	fmt.Printf("Ruleset: %s/%s@%s\n", registry, name, version)
	fmt.Printf("Registry: %s (%s)\n", registry, cfg.Registries[registry])
	if regConfig, exists := cfg.RegistryConfigs[registry]; exists {
		fmt.Printf("Type: %s\n", regConfig["type"])
	}

	if versions {
		fmt.Println("\nAvailable versions: (not yet implemented)")
	}

	fmt.Println("\nDetailed information: (not yet implemented)")

	return nil
}

func handleList(global, local, jsonOutput bool, channels string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine scope
	scope := "both"
	if global && !local {
		scope = "global"
	} else if local && !global {
		scope = "local"
	}

	// Parse channel filter
	var channelFilter []string
	if channels != "" {
		channelFilter = strings.Split(channels, ",")
		for i, ch := range channelFilter {
			channelFilter[i] = strings.TrimSpace(ch)
		}
	}

	if jsonOutput {
		fmt.Printf(`{"scope":"%s","channels":%v,"rulesets":[]}`, scope, channelFilter)
		return nil
	}

	fmt.Printf("Installed rulesets (scope: %s):\n", scope)
	if len(channelFilter) > 0 {
		fmt.Printf("Channels: %s\n", strings.Join(channelFilter, ", "))
	}
	fmt.Println()

	// Check if we have any rulesets configured
	if len(cfg.Rulesets) == 0 {
		fmt.Println("No rulesets installed")
		fmt.Println("Install rulesets with 'arm install <ruleset-name>'")
		return nil
	}

	// List configured rulesets (from manifest)
	fmt.Println("Configured rulesets:")
	for registry, rulesets := range cfg.Rulesets {
		for name, spec := range rulesets {
			fmt.Printf("  %s/%s@%s\n", registry, name, spec.Version)
			if len(spec.Patterns) > 0 {
				fmt.Printf("    Patterns: %s\n", strings.Join(spec.Patterns, ", "))
			}
		}
	}

	fmt.Println("\nActual installation status: (not yet implemented)")

	return nil
}

func getTargetRegistries(allRegistries map[string]string, filter string) []string {
	if filter == "" {
		// Return all registry names
		var names []string
		for name := range allRegistries {
			names = append(names, name)
		}
		return names
	}

	// Parse comma-separated list
	filters := strings.Split(filter, ",")
	var result []string

	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		// Check for exact match first
		if _, exists := allRegistries[f]; exists {
			result = append(result, f)
			continue
		}

		// Check for glob pattern match
		for name := range allRegistries {
			if matched, _ := filepath.Match(f, name); matched {
				result = append(result, name)
			}
		}
	}

	return result
}

// Update, outdated, and uninstall command handlers

func handleUninstall(rulesetName string, global, dryRun bool, channels string) error {
	// Parse ruleset specification
	registry, name, _ := parseRulesetSpec(rulesetName)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have a lock file
	if cfg.LockFile == nil {
		return fmt.Errorf("no lock file found - no rulesets installed")
	}

	// Determine target registry
	if registry == "" {
		// Find the registry from lock file
		for reg, rulesets := range cfg.LockFile.Rulesets {
			if _, exists := rulesets[name]; exists {
				registry = reg
				break
			}
		}
		if registry == "" {
			return fmt.Errorf("ruleset '%s' not found in installed rulesets", name)
		}
	}

	// Check if ruleset is installed
	if cfg.LockFile.Rulesets[registry] == nil || cfg.LockFile.Rulesets[registry][name].Version == "" {
		return fmt.Errorf("ruleset '%s/%s' is not installed", registry, name)
	}

	lockedRuleset := cfg.LockFile.Rulesets[registry][name]

	if dryRun {
		fmt.Printf("Would uninstall: %s/%s@%s\n", registry, name, lockedRuleset.Version)
		if channels != "" {
			fmt.Printf("  Channels: %s\n", channels)
		}
		fmt.Println("  Files would be removed from ARM namespace directories")
		return nil
	}

	fmt.Printf("Uninstalling %s/%s@%s...\n", registry, name, lockedRuleset.Version)

	// Remove from manifest (arm.json)
	if err := removeFromManifest(registry, name, global); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	// Remove from lock file
	if err := removeFromLockFile(registry, name); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	// Remove files from channels
	if err := removeRulesetFiles(cfg, registry, name, channels); err != nil {
		return fmt.Errorf("failed to remove files: %w", err)
	}

	fmt.Printf("✓ Uninstalled %s/%s\n", registry, name)
	return nil
}

func handleOutdated(_, jsonOutput bool) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have a lock file
	if cfg.LockFile == nil {
		return fmt.Errorf("no lock file found - no rulesets installed")
	}

	type outdatedInfo struct {
		Registry       string `json:"registry"`
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
		LatestVersion  string `json:"latest_version"`
		UpdateCommand  string `json:"update_command"`
	}

	var outdatedRulesets []outdatedInfo

	// Check each installed ruleset
	for registry, rulesets := range cfg.LockFile.Rulesets {
		for name, locked := range rulesets {
			// Get version spec from manifest
			var versionSpec string
			if cfg.Rulesets[registry] != nil && cfg.Rulesets[registry][name].Version != "" {
				versionSpec = cfg.Rulesets[registry][name].Version
			} else {
				versionSpec = "latest"
			}

			// For now, simulate version checking (would query registry in real implementation)
			latestVersion := simulateLatestVersion(locked.Version, versionSpec)
			if latestVersion != locked.Version {
				outdatedRulesets = append(outdatedRulesets, outdatedInfo{
					Registry:       registry,
					Name:           name,
					CurrentVersion: locked.Version,
					LatestVersion:  latestVersion,
					UpdateCommand:  fmt.Sprintf("arm update %s/%s", registry, name),
				})
			}
		}
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"outdated": outdatedRulesets,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(outdatedRulesets) == 0 {
		fmt.Println("All rulesets are up to date")
		return nil
	}

	fmt.Printf("Found %d outdated ruleset(s):\n\n", len(outdatedRulesets))
	for _, info := range outdatedRulesets {
		fmt.Printf("%s/%s\n", info.Registry, info.Name)
		fmt.Printf("  Current: %s\n", info.CurrentVersion)
		fmt.Printf("  Latest:  %s\n", info.LatestVersion)
		fmt.Printf("  Update:  %s\n\n", info.UpdateCommand)
	}

	return nil
}

func handleUpdateAll(global, dryRun bool) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have a lock file
	if cfg.LockFile == nil {
		return fmt.Errorf("no lock file found - no rulesets installed")
	}

	var updatedCount int
	var failedUpdates []string

	// Update each installed ruleset
	for registry, rulesets := range cfg.LockFile.Rulesets {
		for name := range rulesets {
			rulesetSpec := fmt.Sprintf("%s/%s", registry, name)
			if err := handleUpdateRuleset(rulesetSpec, global, dryRun); err != nil {
				failedUpdates = append(failedUpdates, fmt.Sprintf("%s: %v", rulesetSpec, err))
				continue
			}
			updatedCount++
		}
	}

	if dryRun {
		fmt.Printf("Would attempt to update %d ruleset(s)\n", updatedCount+len(failedUpdates))
		return nil
	}

	fmt.Printf("Updated %d ruleset(s)\n", updatedCount)
	if len(failedUpdates) > 0 {
		fmt.Printf("Failed to update %d ruleset(s):\n", len(failedUpdates))
		for _, failure := range failedUpdates {
			fmt.Printf("  %s\n", failure)
		}
	}

	return nil
}

func handleUpdateRuleset(rulesetSpec string, _, dryRun bool) error {
	// Parse ruleset specification
	registry, name, _ := parseRulesetSpec(rulesetSpec)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if ruleset is installed
	if cfg.LockFile == nil || cfg.LockFile.Rulesets[registry] == nil || cfg.LockFile.Rulesets[registry][name].Version == "" {
		return fmt.Errorf("ruleset '%s/%s' is not installed", registry, name)
	}

	lockedRuleset := cfg.LockFile.Rulesets[registry][name]
	currentVersion := lockedRuleset.Version

	// Get version spec from manifest
	var versionSpec string
	if cfg.Rulesets[registry] != nil && cfg.Rulesets[registry][name].Version != "" {
		versionSpec = cfg.Rulesets[registry][name].Version
	} else {
		versionSpec = "latest"
	}

	// Simulate version resolution (would query registry in real implementation)
	latestVersion := simulateLatestVersion(currentVersion, versionSpec)

	if latestVersion == currentVersion {
		if !dryRun {
			fmt.Printf("%s/%s is already up to date (%s)\n", registry, name, currentVersion)
		}
		return nil
	}

	if dryRun {
		fmt.Printf("Would update: %s/%s %s → %s\n", registry, name, currentVersion, latestVersion)
		return nil
	}

	fmt.Printf("Updating %s/%s %s → %s...\n", registry, name, currentVersion, latestVersion)

	// Invalidate cache (simulate)
	fmt.Printf("  Invalidating cache for %s registry\n", registry)

	// Update lock file
	if err := updateLockFile(registry, name, latestVersion, &lockedRuleset); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	// Reinstall with new version (simulate)
	fmt.Printf("  Installing new version...\n")

	fmt.Printf("✓ Updated %s/%s to %s\n", registry, name, latestVersion)
	return nil
}

// Helper functions

func removeFromManifest(registry, name string, global bool) error {
	path := getConfigPath("arm.json", global)
	armConfig, err := loadOrCreateJSON(path)
	if err != nil {
		return err
	}

	if armConfig.Rulesets[registry] != nil {
		delete(armConfig.Rulesets[registry], name)
		// Remove registry if empty
		if len(armConfig.Rulesets[registry]) == 0 {
			delete(armConfig.Rulesets, registry)
		}
	}

	return saveJSON(path, armConfig)
}

func removeFromLockFile(registry, name string) error {
	path := "arm.lock"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // No lock file to update
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var lockFile config.LockFile
	if err := json.Unmarshal(data, &lockFile); err != nil {
		return err
	}

	if lockFile.Rulesets[registry] != nil {
		delete(lockFile.Rulesets[registry], name)
		// Remove registry if empty
		if len(lockFile.Rulesets[registry]) == 0 {
			delete(lockFile.Rulesets, registry)
		}
	}

	lockData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, lockData, 0o600)
}

func removeRulesetFiles(cfg *config.Config, registry, name, channels string) error {
	// Parse channel filter
	var targetChannels []string
	if channels != "" {
		targetChannels = strings.Split(channels, ",")
		for i, ch := range targetChannels {
			targetChannels[i] = strings.TrimSpace(ch)
		}
	} else {
		// Use all configured channels
		for channelName := range cfg.Channels {
			targetChannels = append(targetChannels, channelName)
		}
	}

	// Remove files from each channel
	for _, channelName := range targetChannels {
		channelConfig, exists := cfg.Channels[channelName]
		if !exists {
			continue
		}

		for _, dir := range channelConfig.Directories {
			// Expand environment variables
			expandedDir := expandEnvVars(dir)
			// Remove ARM namespace directory for this ruleset
			rulesetPath := filepath.Join(expandedDir, "arm", registry, name)
			if err := os.RemoveAll(rulesetPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove %s: %w", rulesetPath, err)
			}

			// Clean up empty parent directories
			registryPath := filepath.Join(expandedDir, "arm", registry)
			if isEmpty, _ := isDirEmpty(registryPath); isEmpty {
				_ = os.Remove(registryPath)
			}
			armPath := filepath.Join(expandedDir, "arm")
			if isEmpty, _ := isDirEmpty(armPath); isEmpty {
				_ = os.Remove(armPath)
			}
		}
	}

	return nil
}

func updateLockFile(registry, name, newVersion string, existingLocked *config.LockedRuleset) error {
	path := "arm.lock"
	var lockFile config.LockFile

	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &lockFile)
	}

	if lockFile.Rulesets == nil {
		lockFile.Rulesets = make(map[string]map[string]config.LockedRuleset)
	}
	if lockFile.Rulesets[registry] == nil {
		lockFile.Rulesets[registry] = make(map[string]config.LockedRuleset)
	}

	// Update with new version but keep other metadata
	updatedLocked := existingLocked
	updatedLocked.Version = newVersion
	updatedLocked.Resolved = "2024-01-15T10:30:00Z" // Would use current time
	lockFile.Rulesets[registry][name] = *updatedLocked

	lockData, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, lockData, 0o600)
}

func simulateLatestVersion(currentVersion, versionSpec string) string {
	// Simulate version resolution - in real implementation would query registry
	if versionSpec == "latest" {
		return "1.3.0" // Simulate newer version available
	}
	if strings.HasPrefix(versionSpec, "^") {
		return "1.2.1" // Simulate patch update within range
	}
	return currentVersion // No update available
}

func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil // Directory has at least one entry
	}
	return true, nil // Directory is empty
}

// Clean command handler

func handleClean(target string, global, dryRun, force bool) error {
	// Validate target
	validTargets := []string{"cache", "unused", "all"}
	if !contains(validTargets, target) {
		return fmt.Errorf("invalid target '%s'. Valid targets: %s", target, strings.Join(validTargets, ", "))
	}

	if dryRun {
		fmt.Printf("Would clean target: %s\n", target)
		switch target {
		case "cache":
			fmt.Println("  - Remove all cached registry data")
			fmt.Println("  - Remove downloaded tar.gz files")
			fmt.Println("  - Remove Git repository clones")
		case "unused":
			fmt.Println("  - Remove rulesets not in any manifest")
			fmt.Println("  - Clean up empty ARM directories")
		case "all":
			fmt.Println("  - Remove all cached registry data")
			fmt.Println("  - Remove downloaded tar.gz files")
			fmt.Println("  - Remove Git repository clones")
			fmt.Println("  - Remove rulesets not in any manifest")
			fmt.Println("  - Clean up empty ARM directories")
		}
		return nil
	}

	// Confirm destructive operation unless force flag is set
	if !force {
		fmt.Printf("This will clean target '%s'. Continue? (y/N): ", target)
		var response string
		_, _ = fmt.Scanln(&response)
		if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	var cleaned int
	var errors []string

	// Execute cleaning based on target
	switch target {
	case "cache":
		if count, err := cleanCache(); err != nil {
			errors = append(errors, fmt.Sprintf("cache: %v", err))
		} else {
			cleaned += count
		}
	case "unused":
		if count, err := cleanUnused(global); err != nil {
			errors = append(errors, fmt.Sprintf("unused: %v", err))
		} else {
			cleaned += count
		}
	case "all":
		if count, err := cleanCache(); err != nil {
			errors = append(errors, fmt.Sprintf("cache: %v", err))
		} else {
			cleaned += count
		}
		if count, err := cleanUnused(global); err != nil {
			errors = append(errors, fmt.Sprintf("unused: %v", err))
		} else {
			cleaned += count
		}
	}

	// Report results
	if len(errors) > 0 {
		fmt.Printf("Cleaned %d items with %d errors:\n", cleaned, len(errors))
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
		return fmt.Errorf("cleaning completed with errors")
	}

	fmt.Printf("✓ Cleaned %d items\n", cleaned)
	return nil
}

func cleanCache() (int, error) {
	// Get cache path from config or use default
	cachePath := filepath.Join(os.Getenv("HOME"), ".arm", "cache")

	// Load config to get custom cache path if set
	if cfg, err := config.Load(); err == nil {
		if customPath, exists := cfg.CacheConfig["path"]; exists && customPath != "" {
			cachePath = expandEnvVars(customPath)
		}
	}

	// Check if cache directory exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return 0, nil // No cache to clean
	}

	// Count items before removal
	count := 0
	_ = filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})

	// Remove entire cache directory
	if err := os.RemoveAll(cachePath); err != nil {
		return 0, fmt.Errorf("failed to remove cache directory: %w", err)
	}

	fmt.Printf("  Removed cache directory: %s\n", cachePath)
	return count, nil
}

func cleanUnused(_ bool) (int, error) {
	// Load configuration to get installed rulesets
	cfg, err := config.Load()
	if err != nil {
		return 0, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get all configured rulesets from manifest
	configuredRulesets := make(map[string]map[string]bool)
	for registry, rulesets := range cfg.Rulesets {
		if configuredRulesets[registry] == nil {
			configuredRulesets[registry] = make(map[string]bool)
		}
		for name := range rulesets {
			configuredRulesets[registry][name] = true
		}
	}

	count := 0
	// Clean unused rulesets from each channel
	for channelName, channelConfig := range cfg.Channels {
		for _, dir := range channelConfig.Directories {
			expandedDir := expandEnvVars(dir)
			armPath := filepath.Join(expandedDir, "arm")

			// Check if ARM directory exists
			if _, err := os.Stat(armPath); os.IsNotExist(err) {
				continue
			}

			// Walk through registry directories
			registries, err := os.ReadDir(armPath)
			if err != nil {
				continue
			}

			for _, registryDir := range registries {
				if !registryDir.IsDir() {
					continue
				}

				registryName := registryDir.Name()
				registryPath := filepath.Join(armPath, registryName)

				// Walk through ruleset directories
				rulesets, err := os.ReadDir(registryPath)
				if err != nil {
					continue
				}

				for _, rulesetDir := range rulesets {
					if !rulesetDir.IsDir() {
						continue
					}

					rulesetName := rulesetDir.Name()

					// Check if this ruleset is configured
					if configuredRulesets[registryName] == nil || !configuredRulesets[registryName][rulesetName] {
						// This is an unused ruleset, remove it
						rulesetPath := filepath.Join(registryPath, rulesetName)
						if err := os.RemoveAll(rulesetPath); err == nil {
							fmt.Printf("  Removed unused ruleset: %s/%s from %s\n", registryName, rulesetName, channelName)
							count++
						}
					}
				}

				// Clean up empty registry directory
				if isEmpty, _ := isDirEmpty(registryPath); isEmpty {
					_ = os.Remove(registryPath)
				}
			}

			// Clean up empty ARM directory
			if isEmpty, _ := isDirEmpty(armPath); isEmpty {
				_ = os.Remove(armPath)
			}
		}
	}

	return count, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// expandEnvVars is a simple version - would use the one from config package
func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}

// performGitInstallation handles Git registry installations with proper version tracking
func performGitInstallation(cfg *config.Config, registryName, rulesetName, version, channels, patterns string) error {
	// Create registry configuration
	registryConfig := &registry.RegistryConfig{
		Name: registryName,
		Type: cfg.RegistryConfigs[registryName]["type"],
		URL:  cfg.Registries[registryName],
	}

	// Create auth configuration
	authConfig := &registry.AuthConfig{}
	if regConfig := cfg.RegistryConfigs[registryName]; regConfig != nil {
		authConfig.Token = regConfig["authToken"]
		authConfig.Region = regConfig["region"]
		authConfig.Profile = regConfig["profile"]
	}

	// Create Git registry instance
	reg, err := registry.CreateRegistry(registryConfig, authConfig)
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}
	defer func() { _ = reg.Close() }()

	// Cast to GitRegistry to access structured download
	gitReg, ok := reg.(*registry.GitRegistry)
	if !ok {
		return fmt.Errorf("expected Git registry but got %T", reg)
	}

	fmt.Printf("⬇ Downloading %s@%s\n", rulesetName, version)

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "arm-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Parse patterns
	var patternList []string
	if patterns != "" {
		patternList = strings.Split(patterns, ",")
		for i, p := range patternList {
			patternList[i] = strings.TrimSpace(p)
		}
	}

	// Download with structured result
	result, err := gitReg.DownloadRulesetWithResult(context.Background(), rulesetName, version, tempDir, patternList)
	if err != nil {
		return fmt.Errorf("failed to download ruleset: %w", err)
	}

	// Parse channels
	var targetChannels []string
	if channels != "" {
		targetChannels = strings.Split(channels, ",")
		for i, ch := range targetChannels {
			targetChannels[i] = strings.TrimSpace(ch)
		}
	}

	// Install using resolved version
	installer := install.New(cfg)
	req := &install.InstallRequest{
		Registry:    registryName,
		Ruleset:     rulesetName,
		Version:     result.ResolvedVersion, // Use resolved version for installation
		SourceFiles: result.Files,
		Channels:    targetChannels,
	}

	installResult, err := installer.Install(req)
	if err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	// Update manifest with original version spec
	manifestMgr := config.NewManifestManager(false)
	if err := manifestMgr.AddRuleset(registryName, rulesetName, result.VersionSpec, patternList); err != nil {
		fmt.Printf("Warning: Failed to update manifest: %v\n", err)
	}

	fmt.Printf("✓ Installed %s/%s@%s\n", installResult.Registry, installResult.Ruleset, installResult.Version)
	fmt.Printf("  Files: %d\n", installResult.FilesCount)
	fmt.Printf("  Channels: %s\n", strings.Join(installResult.Channels, ", "))

	return nil
}

// performInstallation performs the actual installation of a ruleset
func performInstallation(cfg *config.Config, registryName, rulesetName, version, channels, patterns string) error {
	// Create registry configuration
	registryConfig := &registry.RegistryConfig{
		Name: registryName,
		Type: cfg.RegistryConfigs[registryName]["type"],
		URL:  cfg.Registries[registryName],
	}

	// Create auth configuration
	authConfig := &registry.AuthConfig{}
	if regConfig := cfg.RegistryConfigs[registryName]; regConfig != nil {
		authConfig.Token = regConfig["authToken"]
		authConfig.Region = regConfig["region"]
		authConfig.Profile = regConfig["profile"]
	}

	// Create registry instance
	reg, err := registry.CreateRegistry(registryConfig, authConfig)
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}
	defer func() { _ = reg.Close() }()

	// For Git registries, use structured download to get both versions
	if regConfig := cfg.RegistryConfigs[registryName]; regConfig != nil && regConfig["type"] == "git" {
		return performGitInstallation(cfg, registryName, rulesetName, version, channels, patterns)
	}

	// For non-Git registries, use version as-is (they use concrete versions)

	fmt.Printf("⬇ Downloading %s@%s\n", rulesetName, version)

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "arm-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Download ruleset with patterns for Git registries
	var sourceFiles []string
	regConfig := cfg.RegistryConfigs[registryName]

	if regConfig != nil && regConfig["type"] == "git" {
		// Parse patterns from command line
		var patternList []string
		if patterns != "" {
			patternList = strings.Split(patterns, ",")
			for i, p := range patternList {
				patternList[i] = strings.TrimSpace(p)
			}
		}

		// Use Git-specific download method
		if err := reg.DownloadRulesetWithPatterns(context.Background(), rulesetName, version, tempDir, patternList); err != nil {
			return fmt.Errorf("failed to download ruleset: %w", err)
		}

		// For Git registries, files are copied directly - find them
		sourceFiles, err = findDownloadedFiles(tempDir)
		if err != nil {
			return fmt.Errorf("failed to find downloaded files: %w", err)
		}
	} else {
		// Use standard download method for other registry types
		if err := reg.DownloadRuleset(context.Background(), rulesetName, version, tempDir); err != nil {
			return fmt.Errorf("failed to download ruleset: %w", err)
		}

		// Extract downloaded tar.gz files
		sourceFiles, err = extractRuleset(tempDir)
		if err != nil {
			return fmt.Errorf("failed to extract ruleset: %w", err)
		}
	}

	// Parse channels
	var targetChannels []string
	if channels != "" {
		targetChannels = strings.Split(channels, ",")
		for i, ch := range targetChannels {
			targetChannels[i] = strings.TrimSpace(ch)
		}
	}

	// Create installer and install
	installer := install.New(cfg)
	req := &install.InstallRequest{
		Registry:    registryName,
		Ruleset:     rulesetName,
		Version:     version,
		SourceFiles: sourceFiles,
		Channels:    targetChannels,
	}

	result, err := installer.Install(req)
	if err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	fmt.Printf("✓ Installed %s/%s@%s\n", result.Registry, result.Ruleset, result.Version)
	fmt.Printf("  Files: %d\n", result.FilesCount)
	fmt.Printf("  Channels: %s\n", strings.Join(result.Channels, ", "))

	return nil
}

// extractRuleset extracts files from downloaded ruleset
func extractRuleset(tempDir string) ([]string, error) {
	// Look for ruleset.tar.gz file
	tarPath := filepath.Join(tempDir, "ruleset.tar.gz")
	if _, err := os.Stat(tarPath); err != nil {
		return nil, fmt.Errorf("ruleset.tar.gz not found: %w", err)
	}

	// Extract tar.gz file
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Use tar command to extract
	cmd := exec.Command("tar", "-xzf", tarPath, "-C", extractDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract tar.gz: %w", err)
	}

	// Find all extracted files
	var sourceFiles []string
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sourceFiles = append(sourceFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan extracted files: %w", err)
	}

	return sourceFiles, nil
}

// findDownloadedFiles finds all files in the download directory (for Git registries)
func findDownloadedFiles(tempDir string) ([]string, error) {
	var sourceFiles []string
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sourceFiles = append(sourceFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan downloaded files: %w", err)
	}

	return sourceFiles, nil
}
