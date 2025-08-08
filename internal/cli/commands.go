package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/config"
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
func newConfigCommand(cfg *config.Config) *cobra.Command {
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
	addRegistryCmd.MarkFlagRequired("type")
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
	addChannelCmd.MarkFlagRequired("directories")
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
func newInstallCommand(cfg *config.Config) *cobra.Command {
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
	return &cobra.Command{
		Use:   "uninstall <ruleset-name>",
		Short: "Remove rulesets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("uninstall not implemented")
		},
	}
}

// newSearchCommand creates the search command
func newSearchCommand(cfg *config.Config) *cobra.Command {
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
func newInfoCommand(cfg *config.Config) *cobra.Command {
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
	return &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("outdated not implemented")
		},
	}
}

// newUpdateCommand creates the update command
func newUpdateCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "update [ruleset-name]",
		Short: "Update rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("update not implemented")
		},
	}
}

// newCleanCommand creates the clean command
func newCleanCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "clean [target]",
		Short: "Clean cache and unused rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("clean not implemented")
		},
	}
}

// newListCommand creates the list command
func newListCommand(cfg *config.Config) *cobra.Command {
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

func handleConfigGet(key string, global bool) error {
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

func handleConfigList(global bool) error {
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
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		return ini.Empty(), nil
	}
	return ini.Load(path)
}

func loadOrCreateJSON(path string) (*config.ARMConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
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
	return os.WriteFile(path, data, 0600)
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

func handleInstallFromManifest(global, dryRun bool, channels string) error {
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

	// TODO: Implement actual ruleset installation
	return fmt.Errorf("ruleset installation not yet implemented")
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
