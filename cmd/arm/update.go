package main

import (
	"fmt"

	"github.com/jomadu/arm/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [ruleset-name]",
	Short: "Update installed rulesets to newer versions",
	Long: `Update installed rulesets to newer versions while respecting version constraints.

Examples:
  arm update                    # update all rulesets
  arm update typescript-rules   # update specific ruleset
  arm update --dry-run          # show what would be updated`,
	RunE: runUpdate,
}

var dryRun bool

func init() {
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be updated without making changes")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	u, err := updater.New()
	if err != nil {
		return fmt.Errorf("failed to initialize updater: %w", err)
	}

	var rulesetName string
	if len(args) > 0 {
		rulesetName = args[0]
	}

	if dryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made")
		fmt.Println()
	}

	return u.Update(rulesetName, dryRun)
}
