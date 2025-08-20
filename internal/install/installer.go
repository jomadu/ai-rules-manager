package install

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// Installer manages ruleset installation and file operations
type Installer struct {
	config   *config.Config
	lockPath string
	lockMu   sync.Mutex
}

// InstallRequest represents a ruleset installation request
type InstallRequest struct {
	Registry        string
	Ruleset         string
	Version         string
	ResolvedVersion string   // Actual resolved version (e.g., commit hash)
	SourceFiles     []string // Files to install from cache/extraction
	Channels        []string // Target channels (empty = all channels)
	Patterns        []string // Patterns for git registries
}

// InstallResult represents the result of an installation
type InstallResult struct {
	Registry      string
	Ruleset       string
	Version       string
	InstalledPath string
	FilesCount    int
	Channels      []string
}

// New creates a new installer instance
func New(cfg *config.Config) *Installer {
	return &Installer{
		config:   cfg,
		lockPath: "arm.lock",
	}
}

// Install installs a ruleset to configured channels
func (i *Installer) Install(req *InstallRequest) (*InstallResult, error) {
	if req.Registry == "" || req.Ruleset == "" || req.Version == "" {
		return nil, fmt.Errorf("registry, ruleset, and version are required")
	}

	if len(req.SourceFiles) == 0 {
		return nil, fmt.Errorf("no source files provided")
	}

	// Determine target channels
	targetChannels := req.Channels
	if len(targetChannels) == 0 {
		// Install to all configured channels
		for channelName := range i.config.Channels {
			targetChannels = append(targetChannels, channelName)
		}
	}

	if len(targetChannels) == 0 {
		return nil, fmt.Errorf("no channels configured")
	}

	var installedChannels []string
	var totalFiles int

	// Install to each channel
	for _, channelName := range targetChannels {
		channelConfig, exists := i.config.Channels[channelName]
		if !exists {
			return nil, fmt.Errorf("channel '%s' not configured", channelName)
		}

		for _, channelDir := range channelConfig.Directories {
			// Expand environment variables in channel directory
			expandedDir := expandPath(channelDir)

			// Install to this channel directory
			filesCount, err := i.installToChannel(req, expandedDir)
			if err != nil {
				return nil, fmt.Errorf("failed to install to channel '%s' directory '%s': %w", channelName, expandedDir, err)
			}

			totalFiles += filesCount
		}

		installedChannels = append(installedChannels, channelName)
	}

	// Update lock file with resolved version
	resolvedVersion := req.ResolvedVersion
	if resolvedVersion == "" {
		resolvedVersion = req.Version // Fallback to version if no resolved version provided
	}
	if err := i.updateLockFile(req, resolvedVersion); err != nil {
		return nil, fmt.Errorf("failed to update lock file: %w", err)
	}

	return &InstallResult{
		Registry:      req.Registry,
		Ruleset:       req.Ruleset,
		Version:       req.Version,
		InstalledPath: fmt.Sprintf("arm/%s/%s/%s", req.Registry, req.Ruleset, resolvedVersion),
		FilesCount:    totalFiles,
		Channels:      installedChannels,
	}, nil
}

// installToChannel installs files to a specific channel directory
func (i *Installer) installToChannel(req *InstallRequest, channelDir string) (int, error) {
	// Create ARM namespace directory structure
	armDir := filepath.Join(channelDir, "arm")
	registryDir := filepath.Join(armDir, req.Registry)
	rulesetDir := filepath.Join(registryDir, req.Ruleset)
	// Always use resolved version for directory naming
	dirVersion := req.ResolvedVersion
	if dirVersion == "" {
		dirVersion = req.Version // Fallback for non-Git registries
	}
	versionDir := filepath.Join(rulesetDir, dirVersion)

	// Create directories if they don't exist
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return 0, fmt.Errorf("failed to create version directory: %w", err)
	}

	// Remove previous version after successful installation
	defer i.cleanupPreviousVersion(rulesetDir, dirVersion)

	// Copy files to version directory
	filesCount := 0
	for _, sourceFile := range req.SourceFiles {
		// For Git registries, preserve directory structure by using relative path from temp dir
		// For other registries, use just the filename
		var destPath string
		if strings.Contains(sourceFile, string(filepath.Separator)) {
			// Extract relative path from temp directory structure
			// sourceFile format: /tmp/arm-install-xxx/rules-new/python.mdc
			// We want to preserve: rules-new/python.mdc
			parts := strings.Split(sourceFile, string(filepath.Separator))
			// Find the temp directory part and take everything after it
			for i, part := range parts {
				if strings.HasPrefix(part, "arm-install-") && i+1 < len(parts) {
					// Join all parts after the temp directory
					relativePath := filepath.Join(parts[i+1:]...)
					destPath = filepath.Join(versionDir, relativePath)
					break
				}
			}
			// Fallback if pattern not found
			if destPath == "" {
				destPath = filepath.Join(versionDir, filepath.Base(sourceFile))
			}
		} else {
			// Simple filename, no directory structure
			destPath = filepath.Join(versionDir, filepath.Base(sourceFile))
		}

		// Create destination directory if needed
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return 0, fmt.Errorf("failed to create destination directory: %w", err)
		}

		if err := i.copyFile(sourceFile, destPath); err != nil {
			return 0, fmt.Errorf("failed to copy file '%s': %w", sourceFile, err)
		}

		filesCount++
	}

	return filesCount, nil
}

// copyFile copies a file from source to destination with proper permissions
func (i *Installer) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set file permissions (644 - user read/write, group/other read)
	if err := os.Chmod(dst, 0o644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// cleanupPreviousVersion removes previous version directories, keeping only current
func (i *Installer) cleanupPreviousVersion(rulesetDir, currentVersion string) {
	entries, err := os.ReadDir(rulesetDir)
	if err != nil {
		return // Ignore errors during cleanup
	}

	// Remove all version directories except the current one
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != currentVersion {
			versionPath := filepath.Join(rulesetDir, entry.Name())
			_ = os.RemoveAll(versionPath) // Ignore errors during cleanup
		}
	}
}

// Uninstall removes a ruleset from configured channels
func (i *Installer) Uninstall(registry, ruleset string, channels []string) error {
	if registry == "" || ruleset == "" {
		return fmt.Errorf("registry and ruleset are required")
	}

	// Determine target channels
	targetChannels := channels
	if len(targetChannels) == 0 {
		// Uninstall from all configured channels
		for channelName := range i.config.Channels {
			targetChannels = append(targetChannels, channelName)
		}
	}

	// Uninstall from each channel
	for _, channelName := range targetChannels {
		channelConfig, exists := i.config.Channels[channelName]
		if !exists {
			continue // Skip non-existent channels
		}

		for _, channelDir := range channelConfig.Directories {
			expandedDir := expandPath(channelDir)
			rulesetPath := filepath.Join(expandedDir, "arm", registry, ruleset)

			// Remove entire ruleset directory
			if err := os.RemoveAll(rulesetPath); err != nil {
				return fmt.Errorf("failed to remove ruleset from channel '%s': %w", channelName, err)
			}
		}
	}

	// Update lock file
	if err := i.removeLockEntry(registry, ruleset); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	return nil
}

// ListInstalled returns information about installed rulesets
func (i *Installer) ListInstalled(channels []string) (map[string]map[string][]string, error) {
	result := make(map[string]map[string][]string) // channel -> registry -> []rulesets

	// Determine target channels
	targetChannels := channels
	if len(targetChannels) == 0 {
		// List all configured channels
		for channelName := range i.config.Channels {
			targetChannels = append(targetChannels, channelName)
		}
	}

	// Scan each channel
	for _, channelName := range targetChannels {
		channelConfig, exists := i.config.Channels[channelName]
		if !exists {
			continue
		}

		result[channelName] = make(map[string][]string)

		for _, channelDir := range channelConfig.Directories {
			expandedDir := expandPath(channelDir)
			armDir := filepath.Join(expandedDir, "arm")

			// Scan ARM directory for registries
			registries, err := os.ReadDir(armDir)
			if err != nil {
				continue // ARM directory doesn't exist or can't be read
			}

			for _, registryEntry := range registries {
				if !registryEntry.IsDir() {
					continue
				}

				registryName := registryEntry.Name()
				registryPath := filepath.Join(armDir, registryName)

				// Scan registry directory for rulesets
				rulesets, err := os.ReadDir(registryPath)
				if err != nil {
					continue
				}

				for _, rulesetEntry := range rulesets {
					if !rulesetEntry.IsDir() {
						continue
					}

					rulesetName := rulesetEntry.Name()
					result[channelName][registryName] = append(result[channelName][registryName], rulesetName)
				}
			}
		}
	}

	return result, nil
}

// GetLockFile returns the current lock file content
func (i *Installer) GetLockFile() (*config.LockFile, error) {
	return i.loadLockFile()
}

// expandPath expands environment variables and tilde in file paths
func expandPath(path string) string {
	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	// Handle environment variable expansion
	path = os.ExpandEnv(path)

	return path
}

// updateLockFile updates the lock file with a new ruleset entry
func (i *Installer) updateLockFile(req *InstallRequest, resolvedVersion string) error {
	i.lockMu.Lock()
	defer i.lockMu.Unlock()

	lockFile, err := i.loadLockFile()
	if err != nil {
		return err
	}

	// Initialize registry map if needed
	if lockFile.Rulesets[req.Registry] == nil {
		lockFile.Rulesets[req.Registry] = make(map[string]config.LockedRuleset)
	}

	// Get registry config for metadata
	registryConfig := i.config.RegistryConfigs[req.Registry]
	registryType := ""
	region := ""
	if registryConfig != nil {
		registryType = registryConfig["type"]
		region = registryConfig["region"]
	}

	// Update entry
	lockEntry := config.LockedRuleset{
		Version:  req.Version,
		Resolved: resolvedVersion,
		Registry: i.config.Registries[req.Registry],
		Type:     registryType,
		Region:   region,
	}

	// Add patterns for git registries (passed from install request)
	if registryType == "git" && len(req.Patterns) > 0 {
		lockEntry.Patterns = req.Patterns
	}

	lockFile.Rulesets[req.Registry][req.Ruleset] = lockEntry

	return i.saveLockFile(lockFile)
}

// removeLockEntry removes a ruleset entry from the lock file
func (i *Installer) removeLockEntry(registry, ruleset string) error {
	i.lockMu.Lock()
	defer i.lockMu.Unlock()

	lockFile, err := i.loadLockFile()
	if err != nil {
		return err
	}

	// Remove entry if it exists
	if lockFile.Rulesets[registry] != nil {
		delete(lockFile.Rulesets[registry], ruleset)
		// Remove empty registry
		if len(lockFile.Rulesets[registry]) == 0 {
			delete(lockFile.Rulesets, registry)
		}
	}

	return i.saveLockFile(lockFile)
}

// loadLockFile loads the lock file, creating empty one if missing/corrupted
func (i *Installer) loadLockFile() (*config.LockFile, error) {
	// Try to load existing lock file
	if data, err := os.ReadFile(i.lockPath); err == nil {
		var lockFile config.LockFile
		if err := json.Unmarshal(data, &lockFile); err == nil {
			// Ensure rulesets map is initialized
			if lockFile.Rulesets == nil {
				lockFile.Rulesets = make(map[string]map[string]config.LockedRuleset)
			}
			return &lockFile, nil
		}
	}

	// Create new lock file if missing or corrupted
	return &config.LockFile{
		Rulesets: make(map[string]map[string]config.LockedRuleset),
	}, nil
}

// saveLockFile atomically saves the lock file
func (i *Installer) saveLockFile(lockFile *config.LockFile) error {
	data, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	// Atomic write: write to temp file then rename
	tempPath := i.lockPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp lock file: %w", err)
	}

	// Remove existing lock file if it exists (for Windows compatibility)
	_ = os.Remove(i.lockPath)

	if err := os.Rename(tempPath, i.lockPath); err != nil {
		return fmt.Errorf("failed to rename lock file: %w", err)
	}

	return nil
}

// SyncLockFile regenerates lock file from arm.json configuration
func (i *Installer) SyncLockFile() error {
	i.lockMu.Lock()
	defer i.lockMu.Unlock()

	// Create new lock file from current config
	lockFile := &config.LockFile{
		Rulesets: make(map[string]map[string]config.LockedRuleset),
	}

	// Process all rulesets from config
	for registry, rulesets := range i.config.Rulesets {
		lockFile.Rulesets[registry] = make(map[string]config.LockedRuleset)

		for ruleset, spec := range rulesets {
			// Get registry config for metadata
			registryConfig := i.config.RegistryConfigs[registry]
			registryType := ""
			region := ""
			if registryConfig != nil {
				registryType = registryConfig["type"]
				region = registryConfig["region"]
			}

			// Create lock entry with version spec (will be resolved during install)
			lockFile.Rulesets[registry][ruleset] = config.LockedRuleset{
				Version:  spec.Version,
				Resolved: time.Now().Format(time.RFC3339),
				Registry: i.config.Registries[registry],
				Type:     registryType,
				Region:   region,
			}
		}
	}

	return i.saveLockFile(lockFile)
}
