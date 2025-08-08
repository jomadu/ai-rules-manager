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
func newInstallCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "install [ruleset-spec]",
		Short: "Install rulesets",
		Long:  "Install rulesets from configured registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("install not implemented")
		},
	}
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
func newSearchCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for rulesets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("search not implemented")
		},
	}
}

// newInfoCommand creates the info command
func newInfoCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "info <ruleset-spec>",
		Short: "Show ruleset information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("info not implemented")
		},
	}
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
func newListCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed rulesets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("list not implemented")
		},
	}
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
