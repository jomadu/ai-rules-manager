package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/jomadu/arm/pkg/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed rulesets",
	Long: `Display all installed rulesets with their versions and registry sources.

Shows information from the rules.lock file including ruleset names,
installed versions, and source registries.

Examples:
  arm list                    # Show table format
  arm list --format=json      # Show JSON format
  arm list --format=table     # Show table format (default)`,
	RunE: runList,
}

var listFormat string

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table, json)")
	rootCmd.AddCommand(listCmd)
}

type ListEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

func runList(cmd *cobra.Command, args []string) error {
	// Load lock file
	lock, err := types.LoadLockFile("rules.lock")
	if err != nil {
		// Check if file doesn't exist by trying to stat it
		if _, statErr := os.Stat("rules.lock"); os.IsNotExist(statErr) {
			fmt.Println("No lock file found. Run 'arm install' to install dependencies.")
			return nil
		}
		fmt.Println("Lock file corrupted. Run 'arm install' to rebuild.")
		return nil
	}

	// Check if no rulesets installed
	if len(lock.Dependencies) == 0 {
		fmt.Println("No rulesets installed. Run 'arm install <ruleset>' to get started.")
		return nil
	}

	// Convert to list entries and sort by name
	entries := make([]ListEntry, 0, len(lock.Dependencies))
	for name, dep := range lock.Dependencies {
		entries = append(entries, ListEntry{
			Name:    name,
			Version: dep.Version,
			Source:  formatSource(dep.Source),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	// Output in requested format
	switch listFormat {
	case "json":
		return outputJSON(entries)
	case "table":
		return outputTable(entries)
	default:
		return fmt.Errorf("unsupported format: %s", listFormat)
	}
}

func outputTable(entries []ListEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	if _, err := fmt.Fprintln(w, "NAME\tVERSION\tSOURCE"); err != nil {
		return err
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Name, entry.Version, entry.Source); err != nil {
			return err
		}
	}

	return w.Flush()
}

func outputJSON(entries []ListEntry) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

func formatSource(sourceURL string) string {
	// Load registry config to get friendly names
	config, err := types.LoadRegistryConfig()
	if err != nil {
		return sourceURL // fallback to URL if config can't be loaded
	}

	// Find matching source name
	for name, source := range config.Sources {
		if source.URL == sourceURL {
			if name == "default" {
				return "default"
			}
			return fmt.Sprintf("%s (%s)", name, extractDomain(sourceURL))
		}
	}

	// No friendly name found, just return domain
	return extractDomain(sourceURL)
}

func extractDomain(url string) string {
	// Simple domain extraction - remove protocol and path
	if url == "" {
		return url
	}

	// Remove https:// or http://
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}

	// Find first slash and take everything before it
	for i, char := range url {
		if char == '/' {
			return url[:i]
		}
	}

	return url
}
