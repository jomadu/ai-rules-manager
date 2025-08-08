package cli

import (
	"fmt"

	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root ARM command
func NewRootCommand(cfg *config.Config) *cobra.Command {
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
	rootCmd.AddCommand(newVersionCommand())

	return rootCmd
}

// newConfigCommand creates the config command
func newConfigCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ARM configuration",
		Long:  "Configure registries, channels, and other ARM settings",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("config set not implemented")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Get configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("config get not implemented")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("config list not implemented")
		},
	})

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
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show ARM version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("version not implemented")
		},
	}
}
