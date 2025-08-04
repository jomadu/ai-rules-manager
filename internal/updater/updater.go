package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
	"github.com/schollz/progressbar/v3"
)

type Updater struct {
	configManager *config.Manager
	manager       *registry.Manager
	cacheDir      string
}

type UpdateResult struct {
	Name        string
	OldVersion  string
	NewVersion  string
	Status      UpdateStatus
	Error       error
}

type UpdateStatus int

const (
	UpdateSuccess UpdateStatus = iota
	UpdateFailed
	UpdateSkipped
	UpdateNotNeeded
)

// InstalledRuleset represents an installed ruleset from the lock file
type InstalledRuleset struct {
	Name    string
	Version string
	Source  string
}

func New() (*Updater, error) {
	configManager := config.NewManager()
	if err := configManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	manager := registry.NewManager(configManager)

	// Get cache directory
	cacheDir := ".arm/cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Updater{
		configManager: configManager,
		manager:       manager,
		cacheDir:      cacheDir,
	}, nil
}

func (u *Updater) Update(rulesetName string, dryRun bool) error {
	// Load current lock file
	lockFile, err := types.LoadLockFile("rules.lock")
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Convert lock file to installed rulesets
	var rulesToCheck []InstalledRuleset
	for name, dep := range lockFile.Dependencies {
		if rulesetName != "" && name != rulesetName {
			continue
		}
		rulesToCheck = append(rulesToCheck, InstalledRuleset{
			Name:    name,
			Version: dep.Version,
			Source:  dep.Source,
		})
	}

	if rulesetName != "" && len(rulesToCheck) == 0 {
		return fmt.Errorf("ruleset %s is not installed", rulesetName)
	}

	fmt.Printf("Checking %d ruleset(s) for updates...\n", len(rulesToCheck))

	// Check for updates
	updates, err := u.checkForUpdates(rulesToCheck)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Filter to only rulesets that need updates
	var needsUpdate []UpdateResult
	for _, update := range updates {
		if update.Status == UpdateSuccess {
			needsUpdate = append(needsUpdate, update)
		}
	}

	if len(needsUpdate) == 0 {
		fmt.Println("✓ All rulesets are up to date")
		return nil
	}

	// Show what will be updated
	fmt.Printf("\nFound %d update(s):\n", len(needsUpdate))
	for _, update := range needsUpdate {
		fmt.Printf("  %s: %s → %s\n", update.Name, update.OldVersion, update.NewVersion)
	}

	if dryRun {
		fmt.Println("\n🔍 DRY RUN: No changes were made")
		return nil
	}

	// Perform updates
	fmt.Printf("\nUpdating %d ruleset(s)...\n", len(needsUpdate))
	results := u.performUpdates(needsUpdate)

	// Show results
	u.showResults(results)

	return nil
}

func (u *Updater) checkForUpdates(rulesets []InstalledRuleset) ([]UpdateResult, error) {
	var results []UpdateResult

	bar := progressbar.NewOptions(len(rulesets),
		progressbar.OptionSetDescription("Checking for updates"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	for _, ruleset := range rulesets {
		result := u.checkRulesetUpdate(ruleset)
		results = append(results, result)
		_ = bar.Add(1)
	}

	fmt.Println() // New line after progress bar
	return results, nil
}

func (u *Updater) checkRulesetUpdate(ruleset InstalledRuleset) UpdateResult {
	// Load manifest to get constraints
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("failed to load manifest: %w", err),
		}
	}

	// Get constraint from rules.json
	constraint, exists := manifest.Dependencies[ruleset.Name]
	if !exists {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateSkipped,
			Error:      fmt.Errorf("no constraint found in rules.json"),
		}
	}

	// Parse current version
	currentVer, err := version.NewVersion(ruleset.Version)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("invalid current version: %w", err),
		}
	}

	// Parse constraint
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("invalid version constraint: %w", err),
		}
	}

	// Get available versions from registry
	reg, err := u.manager.GetRegistry(ruleset.Source)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("failed to get registry: %w", err),
		}
	}

	versions, err := reg.ListVersions(ruleset.Name)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("failed to list versions: %w", err),
		}
	}

	// Find the latest version that satisfies constraints
	var latestValid *version.Version
	for _, vStr := range versions {
		v, err := version.NewVersion(vStr)
		if err != nil {
			continue // Skip invalid versions
		}

		if constraints.Check(v) && (latestValid == nil || v.GreaterThan(latestValid)) {
			latestValid = v
		}
	}

	if latestValid == nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateSkipped,
			Error:      fmt.Errorf("no valid versions found"),
		}
	}

	// Check if update is needed
	if !latestValid.GreaterThan(currentVer) {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateNotNeeded,
		}
	}

	return UpdateResult{
		Name:       ruleset.Name,
		OldVersion: ruleset.Version,
		NewVersion: latestValid.String(),
		Status:     UpdateSuccess,
	}
}

func (u *Updater) performUpdates(updates []UpdateResult) []UpdateResult {
	var results []UpdateResult

	bar := progressbar.NewOptions(len(updates),
		progressbar.OptionSetDescription("Updating rulesets"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
	)

	for _, update := range updates {
		result := u.performSingleUpdate(update)
		results = append(results, result)
		_ = bar.Add(1)
	}

	fmt.Println() // New line after progress bar
	return results
}

func (u *Updater) performSingleUpdate(update UpdateResult) UpdateResult {
	// Create backup
	backupPath, err := u.createBackup(update.Name, update.OldVersion)
	if err != nil {
		return UpdateResult{
			Name:       update.Name,
			OldVersion: update.OldVersion,
			NewVersion: update.NewVersion,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("backup failed: %w", err),
		}
	}

	// Attempt update - use simple file operations for now
	err = u.installUpdate(update.Name, update.NewVersion)
	if err != nil {
		// Restore backup on failure
		if restoreErr := u.restoreBackup(update.Name, backupPath); restoreErr != nil {
			return UpdateResult{
				Name:       update.Name,
				OldVersion: update.OldVersion,
				NewVersion: update.NewVersion,
				Status:     UpdateFailed,
				Error:      fmt.Errorf("update failed and restore failed: %w (original: %v)", restoreErr, err),
			}
		}
		return UpdateResult{
			Name:       update.Name,
			OldVersion: update.OldVersion,
			NewVersion: update.NewVersion,
			Status:     UpdateFailed,
			Error:      fmt.Errorf("update failed: %w", err),
		}
	}

	// Clean up backup on success
	_ = os.RemoveAll(backupPath)

	return UpdateResult{
		Name:       update.Name,
		OldVersion: update.OldVersion,
		NewVersion: update.NewVersion,
		Status:     UpdateSuccess,
	}
}

func (u *Updater) createBackup(name, version string) (string, error) {
	// Create backup directory
	backupDir := filepath.Join(u.cacheDir, "backups", name, version)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Load manifest to get targets
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return "", fmt.Errorf("failed to load manifest: %w", err)
	}

	// Copy current installation to backup
	for _, target := range manifest.Targets {
		targetPath := filepath.Join(target, "arm", name, version)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			continue // Skip if target doesn't exist
		}

		backupTargetPath := filepath.Join(backupDir, filepath.Base(target))
		if err := u.copyDir(targetPath, backupTargetPath); err != nil {
			return "", fmt.Errorf("failed to backup %s: %w", target, err)
		}
	}

	return backupDir, nil
}

func (u *Updater) restoreBackup(name, backupPath string) error {
	// Load manifest to get targets
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Restore from backup to all targets
	for _, target := range manifest.Targets {
		backupTargetPath := filepath.Join(backupPath, filepath.Base(target))
		if _, err := os.Stat(backupTargetPath); os.IsNotExist(err) {
			continue // Skip if backup doesn't exist for this target
		}

		targetPath := filepath.Join(target, "arm", name)
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove current installation: %w", err)
		}

		if err := u.copyDir(backupTargetPath, targetPath); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	}

	return nil
}

func (u *Updater) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = srcFile.WriteTo(dstFile)
		return err
	})
}

func (u *Updater) showResults(results []UpdateResult) {
	fmt.Println("\nUpdate Results:")
	
	successCount := 0
	failureCount := 0

	for _, result := range results {
		switch result.Status {
		case UpdateSuccess:
			fmt.Printf("✓ %s: %s → %s\n", result.Name, result.OldVersion, result.NewVersion)
			successCount++
		case UpdateFailed:
			fmt.Printf("✗ %s: failed (%v)\n", result.Name, result.Error)
			failureCount++
		}
	}

	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failureCount)
}

// installUpdate performs the actual update installation
func (u *Updater) installUpdate(name, version string) error {
	// Get registry for this ruleset
	registry, err := u.manager.GetRegistryForRuleset(name)
	if err != nil {
		return fmt.Errorf("failed to get registry: %w", err)
	}

	// Download the new version
	reader, err := registry.Download(name, version)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer reader.Close()

	// For now, just simulate the installation
	// TODO: Implement proper extraction and installation
	fmt.Printf("Installing %s@%s...\n", name, version)

	// Update lock file
	lockFile, err := types.LoadLockFile("rules.lock")
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Update the dependency version
	if dep, exists := lockFile.Dependencies[name]; exists {
		dep.Version = version
		lockFile.Dependencies[name] = dep
	}

	// Save updated lock file
	return lockFile.SaveLockFile("rules.lock")
}