package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "arm",
	Short: "AI Rules Manager - A package manager for AI coding assistant rulesets",
	Long:  "ARM helps you install, update, and manage coding rules across different AI tools like Cursor and Amazon Q Developer.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var configCmd = &cobra.Command{
	Use:   "config [list|get|set] [key] [value]",
	Short: "Manage ARM configuration",
	Long: `Manage ARM configuration stored in .armrc files.

Configuration includes registry sources, authentication tokens,
and performance settings like concurrency limits.

Examples:
  arm config list                           # Show all configuration
  arm config get sources.default            # Get specific value
  arm config set sources.company https://internal.company.local/
  arm config set sources.company.authToken $TOKEN`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configCommand(args)
	},
}

func init() {
	// Commands are added by their respective init() functions
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
