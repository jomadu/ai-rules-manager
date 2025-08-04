package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/internal/updater"
	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated [ruleset-name]",
	Short: "Show available updates for installed rulesets",
	Long: `Show available updates for installed rulesets without performing updates.

Examples:
  arm outdated                    # check all rulesets
  arm outdated typescript-rules   # check specific ruleset
  arm outdated --format=json      # JSON output
  arm outdated --only-outdated    # show only outdated rulesets`,
	RunE: runOutdated,
}

var (
	outdatedFormat       string
	outdatedNoColor      bool
	outdatedOnlyOutdated bool
)

func init() {
	outdatedCmd.Flags().StringVar(&outdatedFormat, "format", "table", "output format (table, json)")
	outdatedCmd.Flags().BoolVar(&outdatedNoColor, "no-color", false, "disable colored output")
	outdatedCmd.Flags().BoolVar(&outdatedOnlyOutdated, "only-outdated", false, "show only outdated rulesets")
	rootCmd.AddCommand(outdatedCmd)
}

func runOutdated(cmd *cobra.Command, args []string) error {
	// Initialize components
	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	manager := registry.NewManager(configManager)
	checker := updater.NewChecker(manager)

	// Check specific ruleset or all
	var results []updater.CheckResult
	var err error

	if len(args) > 0 {
		rulesetName := args[0]
		results, err = checkSpecificRuleset(checker, rulesetName)
	} else {
		results, err = checker.CheckAll()
	}

	if err != nil {
		return err
	}

	// Filter if only-outdated flag is set
	if outdatedOnlyOutdated {
		results = filterOutdated(results)
	}

	// Disable colors if requested or not a terminal
	if outdatedNoColor || !isTerminal() {
		color.NoColor = true
	}

	// Output results
	switch outdatedFormat {
	case "json":
		return outputOutdatedJSON(results)
	case "table":
		return outputOutdatedTable(results)
	default:
		return fmt.Errorf("unsupported format: %s", outdatedFormat)
	}
}

func checkSpecificRuleset(checker *updater.Checker, name string) ([]updater.CheckResult, error) {
	// This would need to load the specific ruleset from lock file
	// For now, get all and filter
	results, err := checker.CheckAll()
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		if result.Name == name {
			return []updater.CheckResult{result}, nil
		}
	}

	return nil, fmt.Errorf("ruleset %s is not installed", name)
}

func filterOutdated(results []updater.CheckResult) []updater.CheckResult {
	var filtered []updater.CheckResult
	for _, result := range results {
		if result.Status == updater.CheckOutdated {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func outputOutdatedTable(results []updater.CheckResult) error {
	if len(results) == 0 {
		fmt.Println("No rulesets found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	if _, err := fmt.Fprintln(w, "NAME\tCURRENT\tAVAILABLE\tCONSTRAINT\tSTATUS"); err != nil {
		return err
	}

	outdatedCount := 0
	upToDateCount := 0
	errorCount := 0

	for _, result := range results {
		status := result.Status.String()
		available := result.Available
		if available == "" {
			available = "-"
		}

		// Apply colors
		switch result.Status {
		case updater.CheckOutdated:
			status = color.YellowString(status)
			outdatedCount++
		case updater.CheckUpToDate:
			status = color.GreenString(status)
			upToDateCount++
		case updater.CheckError, updater.CheckNoCompatible:
			status = color.RedString(status)
			if result.Error != nil {
				status = color.RedString(fmt.Sprintf("%s (%s)", status, result.Error.Error()))
			}
			errorCount++
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			result.Name, result.Current, available, result.Constraint, status); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	// Summary
	fmt.Printf("\n%d rulesets checked", len(results))
	if outdatedCount > 0 {
		fmt.Printf(", %s", color.YellowString(fmt.Sprintf("%d update(s) available", outdatedCount)))
	}
	if upToDateCount > 0 {
		fmt.Printf(", %s", color.GreenString(fmt.Sprintf("%d up to date", upToDateCount)))
	}
	if errorCount > 0 {
		fmt.Printf(", %s", color.RedString(fmt.Sprintf("%d error(s)", errorCount)))
	}
	fmt.Println()

	// Exit with code 1 if updates are available
	if outdatedCount > 0 {
		os.Exit(1)
	}

	return nil
}

func outputOutdatedJSON(results []updater.CheckResult) error {
	type JSONResult struct {
		Name       string `json:"name"`
		Current    string `json:"current"`
		Available  string `json:"available,omitempty"`
		Constraint string `json:"constraint"`
		Status     string `json:"status"`
		Error      string `json:"error,omitempty"`
	}

	type JSONOutput struct {
		Rulesets []JSONResult `json:"rulesets"`
		Summary  struct {
			Total    int `json:"total"`
			Outdated int `json:"outdated"`
			UpToDate int `json:"upToDate"`
			Errors   int `json:"errors"`
		} `json:"summary"`
	}

	output := JSONOutput{}
	output.Summary.Total = len(results)

	for _, result := range results {
		jsonResult := JSONResult{
			Name:       result.Name,
			Current:    result.Current,
			Available:  result.Available,
			Constraint: result.Constraint,
			Status:     strings.ToLower(result.Status.String()),
		}

		if result.Error != nil {
			jsonResult.Error = result.Error.Error()
		}

		output.Rulesets = append(output.Rulesets, jsonResult)

		switch result.Status {
		case updater.CheckOutdated:
			output.Summary.Outdated++
		case updater.CheckUpToDate:
			output.Summary.UpToDate++
		case updater.CheckError, updater.CheckNoCompatible:
			output.Summary.Errors++
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Exit with code 1 if updates are available
	if output.Summary.Outdated > 0 {
		os.Exit(1)
	}

	return nil
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
