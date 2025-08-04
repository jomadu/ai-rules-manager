package main

import (
	"github.com/jomadu/arm/internal/uninstaller"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <ruleset>",
	Short: "Remove a ruleset",
	Long: `Remove a ruleset from target directories and update manifest files.

Removes the ruleset from all configured targets (.cursorrules, .amazonq/rules)
and updates both rules.json and rules.lock files.

Examples:
  arm uninstall typescript-rules    # Remove specific ruleset
  arm uninstall company@rules       # Remove from specific registry`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rulesetName := args[0]

		u := uninstaller.New()
		return u.Uninstall(rulesetName)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
