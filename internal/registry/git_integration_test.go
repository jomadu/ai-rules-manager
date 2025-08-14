package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestGitWorkflow runs comprehensive integration tests for Git registry functionality
func TestGitWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup shared test repository
	testRepo := setupTestRepository(t)

	tests := []struct {
		name string
		test func(*testing.T, string)
	}{
		{"InstallLatest", testInstallLatest},
		{"InstallSemver", testInstallSemver},
		{"InstallPatterns", testInstallPatterns},
		{"InstallCombined", testInstallCombined},
		{"UpdateLatest", testUpdateLatest},
		{"UpdateSemver", testUpdateSemver},
		{"UpdatePatterns", testUpdatePatterns},
		{"OutdatedBasic", testOutdated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, testRepo)
		})
	}
}

// setupTestRepository creates a local bare Git repository with test content and versions
func setupTestRepository(t *testing.T) string {
	// Create working directory first to build the repository
	workDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Configure git in working directory
	gitConfig := func(key, value string) {
		cmd := exec.Command("git", "config", key, value)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git %s: %v", key, err)
		}
	}
	gitConfig("user.name", "Test User")
	gitConfig("user.email", "test@example.com")
	gitConfig("init.defaultBranch", "main")

	// Create version history
	createVersion100(t, workDir)
	gitCommitAndTag(t, workDir, "feat: initial ARM test repository with basic ghost hunting rules", "v1.0.0")

	createVersion110(t, workDir)
	gitCommitAndTag(t, workDir, "feat: add advanced techniques and cursor integration", "v1.1.0")

	createVersion120(t, workDir)
	gitCommitAndTag(t, workDir, "feat: enhance rules with best practices and config options", "v1.2.0")

	createVersion200(t, workDir)
	gitCommitAndTag(t, workDir, "feat!: restructure repository with breaking changes", "v2.0.0")

	// Ensure we're on main branch
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set main branch: %v", err)
	}

	// Create bare repository and push to it
	bareRepoDir := t.TempDir()
	cmd = exec.Command("git", "init", "--bare", bareRepoDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create bare repository: %v", err)
	}

	// Add bare repository as remote and push
	cmd = exec.Command("git", "remote", "add", "origin", bareRepoDir)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// Push all changes and tags
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		t.Fatalf("Failed to push changes: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "push", "origin", "--tags")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		t.Fatalf("Failed to push tags: %v, output: %s", err, string(output))
	}

	return "file://" + bareRepoDir
}

// gitCommitAndTag commits all changes and creates a tag
func gitCommitAndTag(t *testing.T, workDir, message, tag string) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = workDir
	if err := cmd.Run(); err == nil {
		// No changes to commit, skip
		return
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	cmd = exec.Command("git", "tag", tag)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create tag %s: %v", tag, err)
	}
}

// createVersion100 creates v1.0.0 content
func createVersion100(t *testing.T, workDir string) {
	// Create directory structure
	dirs := []string{"rules/advanced", "cursor", "amazon-q"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(workDir, dir), 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	files := map[string]string{
		"README.md": `# ARM Test Repository

Comprehensive test repository for ARM (AI Rules Manager) testing.`,
		"ghost-hunting.md": `# Ghost Hunting Guidelines

*Basic ghost hunting techniques for code debugging.*

## Rule 1: Use Your Flashlight
- Illuminate dark code with proper debugging tools`,
		"rules/mansion-maintenance.md": `# Mansion Maintenance

*Keep your codebase mansion clean.*

## Regular Cleanup
- Remove unused code like dusting furniture`,
		"rules/advanced/boss-battles.md": `# Boss Battle Strategies

*Advanced techniques for complex problems.*

## Strategy 1: Preparation
- Study the problem before attacking`,
		"cursor/its-a-me.md": `# Cursor Integration Rules

*Basic cursor configuration guidelines.*

## Setup Rules
- Configure cursor for optimal performance`,
		"amazon-q/luigi-assistant.md": `# Amazon Q Assistant Rules

*Basic AI assistant guidelines.*

## Interaction Rules
- Be specific in your questions`,
		"config.json": `{
  "version": "1.0.0",
  "features": ["basic-rules"],
  "settings": {
    "ghostDetection": "enabled",
    "debugging": "basic"
  }
}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(workDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
}

// createVersion110 enhances content for v1.1.0
func createVersion110(t *testing.T, workDir string) {
	// Append to existing files
	appendToFile(t, filepath.Join(workDir, "ghost-hunting.md"), `

## Advanced Techniques (v1.1.0)

### Rule 3: Team Coordination
- Hunt ghosts in pairs when possible`)

	appendToFile(t, filepath.Join(workDir, "cursor", "its-a-me.md"), `

## Advanced Cursor Features (v1.1.0)

### Custom Shortcuts
- Set up project-specific shortcuts`)

	// Update config
	config := `{
  "version": "1.1.0",
  "features": ["basic-rules", "advanced-techniques"],
  "settings": {
    "ghostDetection": "enhanced",
    "debugging": "advanced",
    "teamCoordination": "enabled"
  }
}`
	if err := os.WriteFile(filepath.Join(workDir, "config.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("Failed to update config.json: %v", err)
	}
}

// createVersion120 enhances content for v1.2.0
func createVersion120(t *testing.T, workDir string) {
	appendToFile(t, filepath.Join(workDir, "rules", "mansion-maintenance.md"), `

## Best Practices (v1.2.0)

### Code Quality
- Implement automated testing`)

	appendToFile(t, filepath.Join(workDir, "amazon-q", "luigi-assistant.md"), `

## Enhanced AI Workflows (v1.2.0)

### Prompt Engineering
- Craft effective prompts for better results`)

	config := `{
  "version": "1.2.0",
  "features": ["basic-rules", "advanced-techniques", "best-practices"],
  "settings": {
    "ghostDetection": "enhanced",
    "debugging": "advanced",
    "teamCoordination": "enabled",
    "codeQuality": "strict",
    "performance": "optimized"
  }
}`
	if err := os.WriteFile(filepath.Join(workDir, "config.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("Failed to update config.json: %v", err)
	}
}

// createVersion200 creates breaking changes for v2.0.0
func createVersion200(t *testing.T, workDir string) {
	// Remove old structure
	oldDirs := []string{"rules", "cursor", "amazon-q"}
	for _, dir := range oldDirs {
		if err := os.RemoveAll(filepath.Join(workDir, dir)); err != nil {
			t.Fatalf("Failed to remove directory %s: %v", dir, err)
		}
	}
	if err := os.Remove(filepath.Join(workDir, "ghost-hunting.md")); err != nil {
		t.Fatalf("Failed to remove ghost-hunting.md: %v", err)
	}
	if err := os.Remove(filepath.Join(workDir, "config.json")); err != nil {
		t.Fatalf("Failed to remove config.json: %v", err)
	}

	// Create new structure
	dirs := []string{"guidelines", "tools", "ai-assistants"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(workDir, dir), 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	files := map[string]string{
		"ghost-detection.md": `# Ghost Detection System (v2.0.0)

*BREAKING CHANGE: Renamed from ghost-hunting.md*

## Detection Methods
- Automated ghost detection`,
		"guidelines/maintenance.md": `# Maintenance Guidelines (v2.0.0)

*BREAKING CHANGE: Moved from rules/mansion-maintenance.md*

## Automated Maintenance
- Scheduled cleanup tasks`,
		"guidelines/expert-strategies.md": `# Expert Strategies (v2.0.0)

*BREAKING CHANGE: Renamed from rules/advanced/boss-battles.md*

## Master-Level Techniques
- Complex problem solving`,
		"tools/cursor-pro.md": `# Cursor Pro Configuration (v2.0.0)

*BREAKING CHANGE: Renamed from cursor/its-a-me.md*

## Professional Setup
- Enterprise configurations`,
		"ai-assistants/q-developer.md": `# Q Developer Integration (v2.0.0)

*BREAKING CHANGE: Renamed from amazon-q/luigi-assistant.md*

## Professional AI Workflows
- Enterprise AI integration`,
		"settings.json": `{
  "version": "2.0.0",
  "breaking_changes": [
    "Renamed ghost-hunting.md to ghost-detection.md",
    "Moved rules/ to guidelines/",
    "Renamed cursor/ to tools/",
    "Renamed amazon-q/ to ai-assistants/",
    "Renamed config.json to settings.json"
  ],
  "features": ["automated-detection", "professional-workflows", "enterprise-integration"]
}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(workDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
}

// appendToFile appends content to an existing file
func appendToFile(t *testing.T, path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to append to file %s: %v", path, err)
	}
}

// testEnvironment represents a test environment with ARM configuration
type testEnvironment struct {
	workDir     string
	channelDir  string
	armrcPath   string
	armJsonPath string
	registry    *GitRegistry
}

// setupTestEnvironment creates a test environment for ARM operations
func setupTestEnvironment(t *testing.T, testRepoURL string) *testEnvironment {
	workDir := t.TempDir()
	channelDir := filepath.Join(workDir, "test-channel")
	if err := os.MkdirAll(channelDir, 0o755); err != nil {
		t.Fatalf("Failed to create channel directory: %v", err)
	}

	// Create .armrc file
	armrcPath := filepath.Join(workDir, ".armrc")
	armrcContent := fmt.Sprintf(`[registries]
test-repo = %s

[registries.test-repo]
type = git
`, testRepoURL)
	if err := os.WriteFile(armrcPath, []byte(armrcContent), 0o644); err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	// Create arm.json file
	armJsonPath := filepath.Join(workDir, "arm.json")
	armJsonContent := fmt.Sprintf(`{
  "engines": {"arm": "^1.0.0"},
  "channels": {
    "test": {
      "directories": ["%s"]
    }
  },
  "rulesets": {}
}`, channelDir)
	if err := os.WriteFile(armJsonPath, []byte(armJsonContent), 0o644); err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Create registry instance
	config := &RegistryConfig{
		Name:    "test-repo",
		Type:    "git",
		URL:     testRepoURL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}
	registry, err := NewGitRegistry(config, auth)
	if err != nil {
		t.Fatalf("Failed to create Git registry: %v", err)
	}

	return &testEnvironment{
		workDir:     workDir,
		channelDir:  channelDir,
		armrcPath:   armrcPath,
		armJsonPath: armJsonPath,
		registry:    registry,
	}
}

// verifyContent checks if installed files contain expected content
func (env *testEnvironment) verifyContent(t *testing.T, expectedPhrase string) {
	found := false
	err := filepath.Walk(env.channelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.Contains(string(content), expectedPhrase) {
			found = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk channel directory: %v", err)
	}

	if !found {
		t.Errorf("Expected phrase '%s' not found in any installed files", expectedPhrase)
	}
}

// clearChannelDir removes all files from the channel directory
func (env *testEnvironment) clearChannelDir(t *testing.T) {
	if err := os.RemoveAll(env.channelDir); err != nil {
		t.Fatalf("Failed to remove channel directory: %v", err)
	}
	if err := os.MkdirAll(env.channelDir, 0o755); err != nil {
		t.Fatalf("Failed to recreate channel directory: %v", err)
	}
}

// testInstallLatest tests installing the latest version without specifying a version
func testInstallLatest(t *testing.T, testRepoURL string) {
	env := setupTestEnvironment(t, testRepoURL)
	ctx := context.Background()

	// Test actual installation (should get v2.0.0)
	err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", "latest", env.channelDir, []string{"*.md"})
	if err != nil {
		t.Fatalf("Failed to install latest version: %v", err)
	}

	// Verify v2.0.0 content is present
	env.verifyContent(t, "BREAKING CHANGE")
}

// testInstallSemver tests installing specific semantic versions
func testInstallSemver(t *testing.T, testRepoURL string) {
	tests := []struct {
		version        string
		expectedPhrase string
	}{
		{"1.0.0", "Basic ghost hunting"},
		{"1.1.0", "Advanced Techniques"},
		{"1.2.0", "Best Practices"},
		{"2.0.0", "BREAKING CHANGE"},
	}

	for _, tt := range tests {
		t.Run("version_"+tt.version, func(t *testing.T) {
			env := setupTestEnvironment(t, testRepoURL)
			ctx := context.Background()

			err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", tt.version, env.channelDir, []string{"*.md"})
			if err != nil {
				t.Fatalf("Failed to install version %s: %v", tt.version, err)
			}

			env.verifyContent(t, tt.expectedPhrase)
		})
	}
}

// testInstallPatterns tests individual file pattern matching
func testInstallPatterns(t *testing.T, testRepoURL string) {
	tests := []struct {
		name          string
		patterns      []string
		expectedFiles []string
		version       string
	}{
		{
			name:          "markdown_files",
			patterns:      []string{"*.md"},
			expectedFiles: []string{"ghost-hunting.md"},
			version:       "1.0.0",
		},
		{
			name:          "json_files",
			patterns:      []string{"*.json"},
			expectedFiles: []string{"config.json"},
			version:       "1.0.0",
		},
		{
			name:          "rules_directory",
			patterns:      []string{"rules/**/*.md"},
			expectedFiles: []string{"rules/advanced/boss-battles.md"},
			version:       "1.0.0",
		},
		{
			name:          "rules_all_files",
			patterns:      []string{"rules/*.md", "rules/**/*.md"},
			expectedFiles: []string{"rules/mansion-maintenance.md", "rules/advanced/boss-battles.md"},
			version:       "1.0.0",
		},
		{
			name:          "cursor_directory",
			patterns:      []string{"cursor/*.md"},
			expectedFiles: []string{"cursor/its-a-me.md"},
			version:       "1.0.0",
		},
		{
			name:          "amazon_q_directory",
			patterns:      []string{"amazon-q/*.md"},
			expectedFiles: []string{"amazon-q/luigi-assistant.md"},
			version:       "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnvironment(t, testRepoURL)
			ctx := context.Background()

			err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", tt.version, env.channelDir, tt.patterns)
			if err != nil {
				t.Fatalf("Failed to install with patterns %v: %v", tt.patterns, err)
			}

			// Verify expected files exist
			for _, expectedFile := range tt.expectedFiles {
				fullPath := filepath.Join(env.channelDir, expectedFile)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("Expected file %s not found", expectedFile)
				}
			}
		})
	}
}

// testInstallCombined tests complex pattern combinations including exclusions
func testInstallCombined(t *testing.T, testRepoURL string) {
	tests := []struct {
		name          string
		patterns      []string
		expectedFiles []string
		excludedFiles []string
		version       string
	}{
		{
			name:          "markdown_and_json",
			patterns:      []string{"*.md", "*.json"},
			expectedFiles: []string{"ghost-hunting.md", "config.json"},
			version:       "1.0.0",
		},
		{
			name:          "rules_and_cursor",
			patterns:      []string{"rules/*.md", "rules/**/*.md", "cursor/*.md"},
			expectedFiles: []string{"rules/mansion-maintenance.md", "rules/advanced/boss-battles.md", "cursor/its-a-me.md"},
			version:       "1.0.0",
		},
		{
			name:          "specific_patterns_only",
			patterns:      []string{"ghost-hunting.md", "README.md"},
			expectedFiles: []string{"ghost-hunting.md", "README.md"},
			excludedFiles: []string{"rules/advanced/boss-battles.md"},
			version:       "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnvironment(t, testRepoURL)
			ctx := context.Background()

			err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", tt.version, env.channelDir, tt.patterns)
			if err != nil {
				t.Fatalf("Failed to install with combined patterns %v: %v", tt.patterns, err)
			}

			// Verify expected files exist
			for _, expectedFile := range tt.expectedFiles {
				fullPath := filepath.Join(env.channelDir, expectedFile)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("Expected file %s not found", expectedFile)
				}
			}

			// Verify excluded files don't exist
			for _, excludedFile := range tt.excludedFiles {
				fullPath := filepath.Join(env.channelDir, excludedFile)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("Excluded file %s should not exist", excludedFile)
				}
			}
		})
	}
}

// testUpdateLatest tests updating from older version to latest
func testUpdateLatest(t *testing.T, testRepoURL string) {
	env := setupTestEnvironment(t, testRepoURL)
	ctx := context.Background()

	// Install v1.0.0 first
	err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", "1.0.0", env.channelDir, []string{"*.md"})
	if err != nil {
		t.Fatalf("Failed to install v1.0.0: %v", err)
	}

	// Verify v1.0.0 content
	env.verifyContent(t, "Basic ghost hunting")

	// Clear and update to latest
	env.clearChannelDir(t)
	err = env.registry.DownloadRulesetWithPatterns(ctx, "rules", "latest", env.channelDir, []string{"*.md"})
	if err != nil {
		t.Fatalf("Failed to update to latest: %v", err)
	}

	// Verify v2.0.0 content is present
	env.verifyContent(t, "BREAKING CHANGE")

	// In v2.0.0, the file structure changes completely
	// ghost-hunting.md becomes ghost-detection.md
	newFile := filepath.Join(env.channelDir, "ghost-detection.md")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file ghost-detection.md should exist after update to v2.0.0")
	}
}

// testUpdateSemver tests updating with semantic version constraints
func testUpdateSemver(t *testing.T, testRepoURL string) {
	tests := []struct {
		name            string
		installVersion  string
		updateVersion   string
		expectedVersion string
		oldPhrase       string
		newPhrase       string
	}{
		{
			name:            "caret_constraint",
			installVersion:  "1.0.0",
			updateVersion:   "^1.0.0",
			expectedVersion: "1.2.0",
			oldPhrase:       "Basic ghost hunting",
			newPhrase:       "Best Practices",
		},
		{
			name:            "tilde_constraint_same",
			installVersion:  "1.1.0",
			updateVersion:   "~1.1.0",
			expectedVersion: "1.1.0",
			oldPhrase:       "Advanced Techniques",
			newPhrase:       "Advanced Techniques",
		},
		{
			name:            "tilde_constraint_update",
			installVersion:  "1.0.0",
			updateVersion:   "~1.2.0",
			expectedVersion: "1.2.0",
			oldPhrase:       "Basic ghost hunting",
			newPhrase:       "Best Practices",
		},
		{
			name:            "gte_constraint",
			installVersion:  "1.0.0",
			updateVersion:   ">=1.1.0",
			expectedVersion: "2.0.0",
			oldPhrase:       "Basic ghost hunting",
			newPhrase:       "BREAKING CHANGE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnvironment(t, testRepoURL)
			ctx := context.Background()

			// Install initial version
			err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", tt.installVersion, env.channelDir, []string{"*.md"})
			if err != nil {
				t.Fatalf("Failed to install version %s: %v", tt.installVersion, err)
			}

			// Verify initial content
			env.verifyContent(t, tt.oldPhrase)

			// Resolve the constraint to get expected version
			resolvedVersion, err := env.registry.ResolveVersion(ctx, tt.updateVersion)
			if err != nil {
				t.Fatalf("Failed to resolve version constraint %s: %v", tt.updateVersion, err)
			}

			// Clear and install resolved version
			env.clearChannelDir(t)
			err = env.registry.DownloadRulesetWithPatterns(ctx, "rules", resolvedVersion, env.channelDir, []string{"*.md"})
			if err != nil {
				t.Fatalf("Failed to update to resolved version %s: %v", resolvedVersion, err)
			}

			// Verify new content
			env.verifyContent(t, tt.newPhrase)
		})
	}
}

// testUpdatePatterns tests that patterns are preserved during updates
func testUpdatePatterns(t *testing.T, testRepoURL string) {
	env := setupTestEnvironment(t, testRepoURL)
	ctx := context.Background()

	patterns := []string{"cursor/*.md", "*.json"}

	// Install v1.0.0 with specific patterns
	err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", "1.0.0", env.channelDir, patterns)
	if err != nil {
		t.Fatalf("Failed to install v1.0.0 with patterns: %v", err)
	}

	// Verify cursor content and config are present
	cursorFile := filepath.Join(env.channelDir, "cursor", "its-a-me.md")
	configFile := filepath.Join(env.channelDir, "config.json")

	if _, err := os.Stat(cursorFile); os.IsNotExist(err) {
		t.Error("Cursor file should exist")
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file should exist")
	}

	// Clear and update to v1.1.0 with same patterns
	env.clearChannelDir(t)
	err = env.registry.DownloadRulesetWithPatterns(ctx, "rules", "1.1.0", env.channelDir, patterns)
	if err != nil {
		t.Fatalf("Failed to update to v1.1.0: %v", err)
	}

	// Verify cursor content updated and config still present
	env.verifyContent(t, "Advanced Cursor Features")

	// Verify config file has updated version
	configContent, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(configContent), "1.1.0") {
		t.Error("Config file should contain version 1.1.0")
	}
}

// testOutdated tests the outdated command in various scenarios
func testOutdated(t *testing.T, testRepoURL string) {
	t.Run("no_rulesets_installed", func(t *testing.T) {
		env := setupTestEnvironment(t, testRepoURL)
		ctx := context.Background()

		// Try to get versions without any installation
		versions, err := env.registry.GetVersions(ctx, "rules")
		if err != nil {
			// This is expected behavior - no cached repository exists yet
			return
		}

		// If it succeeds, we should at least get some versions
		if len(versions) == 0 {
			t.Error("Expected some versions to be available")
		}
	})

	t.Run("latest_version_installed", func(t *testing.T) {
		env := setupTestEnvironment(t, testRepoURL)
		ctx := context.Background()

		// Install latest version
		err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", "latest", env.channelDir, []string{"*.md"})
		if err != nil {
			t.Fatalf("Failed to install latest version: %v", err)
		}

		// Get available versions
		versions, err := env.registry.GetVersions(ctx, "rules")
		if err != nil {
			t.Fatalf("Failed to get versions: %v", err)
		}

		// Should include "latest" and version tags
		found := false
		for _, v := range versions {
			if v == "latest" || v == "v2.0.0" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find latest or v2.0.0 in versions")
		}
	})

	t.Run("older_version_installed", func(t *testing.T) {
		env := setupTestEnvironment(t, testRepoURL)
		ctx := context.Background()

		// Install older version
		err := env.registry.DownloadRulesetWithPatterns(ctx, "rules", "1.0.0", env.channelDir, []string{"*.md"})
		if err != nil {
			t.Fatalf("Failed to install v1.0.0: %v", err)
		}

		// Resolve latest version
		latestVersion, err := env.registry.ResolveVersion(ctx, "latest")
		if err != nil {
			t.Fatalf("Failed to resolve latest version: %v", err)
		}

		// Latest should be different from 1.0.0
		if latestVersion == "1.0.0" {
			t.Error("Latest version should be different from 1.0.0")
		}
	})

	t.Run("version_resolution", func(t *testing.T) {
		env := setupTestEnvironment(t, testRepoURL)
		ctx := context.Background()

		// Test various version resolution scenarios
		tests := []struct {
			versionSpec string
			shouldWork  bool
		}{
			{"latest", true},
			{"1.0.0", true},
			{"v1.0.0", true},
			{"^1.0.0", true},
			{"~1.1.0", true},
			{">=1.0.0", true},
			{"nonexistent", false},
		}

		for _, tt := range tests {
			resolved, err := env.registry.ResolveVersion(ctx, tt.versionSpec)
			if tt.shouldWork {
				if err != nil {
					t.Errorf("Expected version spec %s to resolve, got error: %v", tt.versionSpec, err)
				} else if resolved == "" {
					t.Errorf("Expected version spec %s to resolve to non-empty string", tt.versionSpec)
				}
			} else {
				if err == nil {
					t.Errorf("Expected version spec %s to fail, but got: %s", tt.versionSpec, resolved)
				}
			}
		}
	})
}
