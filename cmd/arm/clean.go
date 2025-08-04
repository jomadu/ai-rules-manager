package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jomadu/arm/internal/cleaner"
	"github.com/jomadu/arm/pkg/types"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean unused rulesets or global cache",
	Long: `Clean unused rulesets from project targets or clear the global cache.

Examples:
  arm clean              # Remove unused rulesets from project targets
  arm clean --dry-run    # Show what would be removed without doing it
  arm clean --cache      # Clear the entire global cache (with confirmation)`,
	RunE: runClean,
}

var (
	cleanCache  bool
	cleanDryRun bool
)

func init() {
	cleanCmd.Flags().BoolVar(&cleanCache, "cache", false, "Clear the entire global cache")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be cleaned without doing it")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	if cleanCache {
		return cleanGlobalCache()
	}
	return cleanProjectTargets()
}

func cleanGlobalCache() error {
	cacheDir, err := types.GetCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Check if cache exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Println("Global cache directory does not exist")
		return nil
	}

	if cleanDryRun {
		fmt.Printf("Would remove entire global cache directory: %s\n", cacheDir)
		return nil
	}

	// Ask for confirmation
	fmt.Printf("This will permanently delete the entire global cache directory:\n%s\n", cacheDir)
	fmt.Print("Are you sure? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cache cleanup cancelled")
		return nil
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	fmt.Println("Global cache cleared successfully")
	return nil
}

func cleanProjectTargets() error {
	// Load manifest to get targets
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		// If no rules.json, use default targets
		manifest = &types.RulesManifest{
			Targets: types.GetDefaultTargets(),
		}
	}

	// Load lock file to get currently used rulesets
	var usedRulesets map[string]bool
	if lock, err := types.LoadLockFile("rules.lock"); err == nil {
		usedRulesets = make(map[string]bool)
		for name := range lock.Dependencies {
			usedRulesets[name] = true
		}
	}

	cleaner := cleaner.New()
	return cleaner.CleanTargets(manifest.Targets, usedRulesets, cleanDryRun)
}
