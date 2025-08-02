package main

import (
	"github.com/jomadu/arm/internal/uninstaller"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <ruleset>",
	Short: "Remove a ruleset",
	Long:  "Remove a ruleset from target directories and update manifest files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rulesetName := args[0]

		u := uninstaller.New()
		return u.Uninstall(rulesetName)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
