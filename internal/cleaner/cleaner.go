package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Cleaner struct{}

func New() *Cleaner {
	return &Cleaner{}
}

func (c *Cleaner) CleanTargets(targets []string, usedRulesets map[string]bool, dryRun bool) error {
	var totalRemoved int
	var errors []string

	for _, target := range targets {
		removed, err := c.cleanTarget(target, usedRulesets, dryRun)
		if err != nil {
			errors = append(errors, fmt.Sprintf("target %s: %v", target, err))
			continue
		}
		totalRemoved += removed
	}

	if len(errors) > 0 {
		fmt.Printf("Errors occurred during cleanup:\n")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if dryRun {
		if totalRemoved > 0 {
			fmt.Printf("Would remove %d unused ruleset(s)\n", totalRemoved)
		} else {
			fmt.Println("No unused rulesets found")
		}
	} else {
		if totalRemoved > 0 {
			fmt.Printf("Removed %d unused ruleset(s)\n", totalRemoved)
		} else {
			fmt.Println("No unused rulesets found")
		}
	}

	return nil
}

func (c *Cleaner) cleanTarget(target string, usedRulesets map[string]bool, dryRun bool) (int, error) {
	armDir := filepath.Join(target, "arm")

	// Check if arm directory exists
	if _, err := os.Stat(armDir); os.IsNotExist(err) {
		return 0, nil // No arm directory, nothing to clean
	}

	var removed int
	err := filepath.Walk(armDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the arm directory itself
		if path == armDir {
			return nil
		}

		// Look for ruleset directories (could be org/package or just package)
		if info.IsDir() {
			rulesetName := c.extractRulesetName(armDir, path)
			if rulesetName != "" && !usedRulesets[rulesetName] {
				// Only remove if this is a complete ruleset (has version subdirectories)
				if c.isRulesetDirectory(path) {
					if dryRun {
						fmt.Printf("Would remove: %s\n", path)
					} else {
						if err := os.RemoveAll(path); err != nil {
							return fmt.Errorf("failed to remove %s: %w", path, err)
						}
						fmt.Printf("Removed: %s\n", path)
					}
					removed++
					return filepath.SkipDir // Skip walking into this directory
				}
			}
		}

		return nil
	})

	if err != nil {
		return removed, err
	}

	// Clean up empty directories
	if !dryRun {
		c.cleanEmptyDirs(armDir)
	}

	return removed, nil
}

func (c *Cleaner) extractRulesetName(armDir, path string) string {
	// Get relative path from arm directory
	relPath, err := filepath.Rel(armDir, path)
	if err != nil || relPath == "." {
		return ""
	}

	parts := strings.Split(relPath, string(filepath.Separator))

	// Handle org/package structure (@org/package)
	if len(parts) == 2 && strings.HasPrefix(parts[0], "@") {
		org := strings.TrimPrefix(parts[0], "@")
		pkg := parts[1]
		return org + "@" + pkg
	}

	// Handle simple package structure (only direct children of arm/)
	if len(parts) == 1 && parts[0] != "." && !strings.HasPrefix(parts[0], "@") {
		return parts[0]
	}

	return ""
}

func (c *Cleaner) isRulesetDirectory(path string) bool {
	// Check if this directory contains version subdirectories or rule files
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	// If it has subdirectories that look like versions, it's a ruleset
	for _, entry := range entries {
		if entry.IsDir() {
			return true // Assume version directories
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			return true // Has rule files
		}
	}

	return false
}

func (c *Cleaner) cleanEmptyDirs(dir string) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == dir {
			return err
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			_ = os.Remove(path) // Ignore errors for empty dir cleanup
		}

		return nil
	})
}
