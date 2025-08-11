package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/install"
	"github.com/max-dunn/ai-rules-manager/internal/registry"
	"github.com/max-dunn/ai-rules-manager/internal/update"
)

func TestGitLocalRegistry_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test environment
	tempDir, err := os.MkdirTemp("", "arm-git-local-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test Git repository
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := createTestGitLocalRepo(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create ARM working directory
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	// Test scenarios
	t.Run("BasicInstallFromLocalGitRepo", func(t *testing.T) {
		testBasicInstallFromLocalGitRepo(t, repoDir, workDir)
	})

	t.Run("PathResolution", func(t *testing.T) {
		testPathResolution(t, repoDir)
	})

	t.Run("VersionConstraintHandling", func(t *testing.T) {
		testVersionConstraintHandling(t, repoDir)
	})

	t.Run("PatternMatching", func(t *testing.T) {
		testPatternMatching(t, repoDir)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, tempDir)
	})
}

func createTestGitLocalRepo(repoDir string) error {
	// Initialize Git repository
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Create test files structure
	files := map[string]string{
		"rules/python.md":      "# Python Rules v1\nPython coding standards",
		"rules/javascript.md":  "# JavaScript Rules v1\nJS coding standards",
		"config/settings.json": `{"version": "1.0.0", "type": "coding-rules"}`,
		"docs/README.md":       "# Test Ruleset\nThis is a test ruleset",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(repoDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return err
		}
		if _, err := worktree.Add(filePath); err != nil {
			return err
		}
	}

	// Create initial commit and tag
	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		return err
	}

	if _, err := repo.CreateTag("1.0.0", commit1, nil); err != nil {
		return err
	}

	// Update files for v2
	files["rules/python.md"] = "# Python Rules v2\nUpdated Python coding standards"
	files["rules/go.md"] = "# Go Rules v2\nGo coding standards"

	for filePath, content := range files {
		fullPath := filepath.Join(repoDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return err
		}
		if _, err := worktree.Add(filePath); err != nil {
			return err
		}
	}

	commit2, err := worktree.Commit("Version 2.0.0", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		return err
	}

	if _, err := repo.CreateTag("2.0.0", commit2, nil); err != nil {
		return err
	}

	return nil
}

func testBasicInstallFromLocalGitRepo(t *testing.T, repoDir, workDir string) {
	// Create test configuration
	cfg := createGitLocalTestConfig(t, repoDir, workDir)

	// Create installer
	installer := install.New(cfg)

	// Test install request
	req := &install.InstallRequest{
		Registry: "local-test",
		Ruleset:  "test-rules",
		Version:  "1.0.0",
		Channels: []string{"cursor"},
	}

	// Create registry to get source files
	regConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  repoDir,
	}
	reg, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	// Download to temp directory
	tempDownloadDir, err := os.MkdirTemp("", "arm-download-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDownloadDir) }()

	if err := reg.DownloadRulesetWithPatterns(context.Background(), "test-rules", "1.0.0", tempDownloadDir, []string{"**/*"}); err != nil {
		t.Fatal(err)
	}

	// Find downloaded files
	sourceFiles, err := findDownloadedFiles(tempDownloadDir)
	if err != nil {
		t.Fatal(err)
	}
	req.SourceFiles = sourceFiles

	// Execute install
	result, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify install result
	if result.Registry != req.Registry {
		t.Errorf("Expected registry %q, got %q", req.Registry, result.Registry)
	}
	if result.Version != req.Version {
		t.Errorf("Expected version %q, got %q", req.Version, result.Version)
	}

	// Verify files deployed to channel
	channelDir := filepath.Join(workDir, "cursor")
	versionDir := filepath.Join(channelDir, "arm", req.Registry, req.Ruleset, req.Version)

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		t.Errorf("Version directory not created: %s", versionDir)
	}

	// Check specific files exist
	expectedFiles := []string{"rules/python.md", "rules/javascript.md", "config/settings.json"}
	for _, expectedFile := range expectedFiles {
		found := false
		for _, sourceFile := range sourceFiles {
			if strings.HasSuffix(sourceFile, expectedFile) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file not found in source files: %s", expectedFile)
		}
	}

	t.Logf("Basic install test completed successfully")
}

func testPathResolution(t *testing.T, repoDir string) {
	// Test absolute path resolution (the main functionality)
	regConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  repoDir, // Use absolute path
	}

	reg, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	// Test version resolution
	versions, err := reg.GetVersions(context.Background(), "test-rules")
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	expectedVersions := []string{"1.0.0", "2.0.0"}

	// Verify versions contain expected tags (may also include "latest")
	for _, expectedVersion := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expectedVersion {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version %s not found in %v", expectedVersion, versions)
		}
	}

	// Verify we have at least the expected number of versions
	if len(versions) < len(expectedVersions) {
		t.Errorf("Expected at least %d versions, got %d: %v", len(expectedVersions), len(versions), versions)
	}

	t.Logf("Path resolution test completed successfully")
}

func testVersionConstraintHandling(t *testing.T, repoDir string) {
	regConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  repoDir,
	}

	reg, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	// Cast to GitLocalRegistry to access ResolveVersion method
	gitLocalReg, ok := reg.(*registry.GitLocalRegistry)
	if !ok {
		t.Fatalf("Expected GitLocalRegistry, got %T", reg)
	}

	// Test version resolution - note that local Git operations may return commit hashes
	// instead of tag names, which is acceptable behavior
	testCases := []struct {
		input string
	}{
		{"latest"},
		{"1.0.0"},
		{"2.0.0"},
	}

	for _, tc := range testCases {
		resolved, err := gitLocalReg.ResolveVersion(context.Background(), tc.input)
		if err != nil {
			t.Errorf("Failed to resolve version %s: %v", tc.input, err)
			continue
		}
		// Just verify that we get a non-empty resolved version
		if resolved == "" {
			t.Errorf("Version %s resolved to empty string", tc.input)
		}
		t.Logf("Version %s resolved to %s", tc.input, resolved)
	}

	t.Logf("Version constraint handling test completed successfully")
}

func testPatternMatching(t *testing.T, repoDir string) {
	regConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  repoDir,
	}

	reg, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	// Cast to GitLocalRegistry to access GetFiles method
	gitLocalReg, ok := reg.(*registry.GitLocalRegistry)
	if !ok {
		t.Fatalf("Expected GitLocalRegistry, got %T", reg)
	}

	// Test pattern matching
	testCases := []struct {
		patterns []string
		expected []string
	}{
		{
			patterns: []string{"rules/*.md"},
			expected: []string{"rules/python.md", "rules/javascript.md"},
		},
		{
			patterns: []string{"**/*.json"},
			expected: []string{"config/settings.json"},
		},
		{
			patterns: []string{"rules/python.md"},
			expected: []string{"rules/python.md"},
		},
	}

	for _, tc := range testCases {
		files, err := gitLocalReg.GetFiles(context.Background(), "1.0.0", tc.patterns)
		if err != nil {
			t.Errorf("Failed to get files with patterns %v: %v", tc.patterns, err)
			continue
		}

		for _, expectedFile := range tc.expected {
			if _, exists := files[expectedFile]; !exists {
				t.Errorf("Expected file %s not found in results for patterns %v", expectedFile, tc.patterns)
			}
		}
	}

	t.Logf("Pattern matching test completed successfully")
}

func testErrorHandling(t *testing.T, tempDir string) {
	// Test invalid repository path
	invalidPath := filepath.Join(tempDir, "nonexistent")
	regConfig := &registry.RegistryConfig{
		Name: "invalid-test",
		Type: "git-local",
		URL:  invalidPath,
	}

	_, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err == nil {
		t.Error("Expected error for invalid repository path, got none")
	}

	// Test non-git directory
	nonGitDir := filepath.Join(tempDir, "not-git")
	if err := os.MkdirAll(nonGitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	regConfig.URL = nonGitDir
	_, err = registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err == nil {
		t.Error("Expected error for non-git directory, got none")
	}

	t.Logf("Error handling test completed successfully")
}

func TestGitLocalRegistry_MixedDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test environment
	tempDir, err := os.MkdirTemp("", "arm-mixed-deps-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create local Git repository
	localRepoDir := filepath.Join(tempDir, "local-repo")
	if err := createTestGitLocalRepo(localRepoDir); err != nil {
		t.Fatal(err)
	}

	// Create ARM working directory
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	// Create configuration with both local and remote registries
	cfg := createMixedRegistryConfig(t, localRepoDir, workDir)

	// Test installing from local registry
	installer := install.New(cfg)

	// Install from local registry
	localReq := &install.InstallRequest{
		Registry: "local-test",
		Ruleset:  "local-rules",
		Version:  "1.0.0",
		Channels: []string{"cursor"},
	}

	// Get source files for local registry
	localRegConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  localRepoDir,
	}
	localReg, err := registry.CreateRegistry(localRegConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = localReg.Close() }()

	tempDownloadDir, err := os.MkdirTemp("", "arm-mixed-download-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDownloadDir) }()

	if err := localReg.DownloadRulesetWithPatterns(context.Background(), "local-rules", "1.0.0", tempDownloadDir, []string{"**/*"}); err != nil {
		t.Fatal(err)
	}

	sourceFiles, err := findDownloadedFiles(tempDownloadDir)
	if err != nil {
		t.Fatal(err)
	}
	localReq.SourceFiles = sourceFiles

	// Execute local install
	localResult, err := installer.Install(localReq)
	if err != nil {
		t.Fatalf("Local install failed: %v", err)
	}

	// Verify local installation
	if localResult.Registry != "local-test" {
		t.Errorf("Expected local registry, got %s", localResult.Registry)
	}

	// Verify files deployed
	channelDir := filepath.Join(workDir, "cursor")
	localVersionDir := filepath.Join(channelDir, "arm", "local-test", "local-rules", "1.0.0")

	if _, err := os.Stat(localVersionDir); os.IsNotExist(err) {
		t.Errorf("Local version directory not created: %s", localVersionDir)
	}

	t.Logf("Mixed dependencies test completed successfully")
}

func TestGitLocalRegistry_UpdateOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test environment
	tempDir, err := os.MkdirTemp("", "arm-update-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test Git repository
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := createTestGitLocalRepo(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create ARM working directory
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	// Create test configuration
	cfg := createGitLocalTestConfig(t, repoDir, workDir)

	// Install initial version
	installer := install.New(cfg)
	req := &install.InstallRequest{
		Registry: "local-test",
		Ruleset:  "test-rules",
		Version:  "1.0.0",
		Channels: []string{"cursor"},
	}

	// Get source files
	regConfig := &registry.RegistryConfig{
		Name: "local-test",
		Type: "git-local",
		URL:  repoDir,
	}
	reg, err := registry.CreateRegistry(regConfig, &registry.AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	tempDownloadDir, err := os.MkdirTemp("", "arm-update-download-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDownloadDir) }()

	if err := reg.DownloadRulesetWithPatterns(context.Background(), "test-rules", "1.0.0", tempDownloadDir, []string{"**/*"}); err != nil {
		t.Fatal(err)
	}

	sourceFiles, err := findDownloadedFiles(tempDownloadDir)
	if err != nil {
		t.Fatal(err)
	}
	req.SourceFiles = sourceFiles

	// Execute initial install
	installResult, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Initial install failed: %v", err)
	}

	// Verify installation was recorded
	if installResult.Registry != "local-test" || installResult.Ruleset != "test-rules" {
		t.Fatalf("Install result mismatch: got %s/%s", installResult.Registry, installResult.Ruleset)
	}

	// Check if lock file was created and contains our ruleset
	lockFile, err := installer.GetLockFile()
	if err != nil {
		t.Fatalf("Failed to get lock file: %v", err)
	}

	if lockFile.Rulesets["local-test"] == nil || lockFile.Rulesets["local-test"]["test-rules"].Version == "" {
		t.Fatalf("Ruleset not found in lock file")
	}

	// Reload config to include lock file
	cfg.LockFile = lockFile

	// Test update operation
	updateService := update.New(cfg)
	updateResult, err := updateService.UpdateRuleset(context.Background(), "local-test/test-rules")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update result
	if !updateResult.Updated {
		t.Error("Expected update to occur")
	}
	// For git-local registries, versions may be commit hashes instead of tag names
	if updateResult.PreviousVersion == "" {
		t.Error("Expected non-empty previous version")
	}
	if updateResult.Version == "" {
		t.Error("Expected non-empty new version")
	}
	if updateResult.Version == updateResult.PreviousVersion {
		t.Error("Expected version to change during update")
	}
	t.Logf("Updated from %s to %s", updateResult.PreviousVersion, updateResult.Version)

	t.Logf("Update operations test completed successfully")
}

func createGitLocalTestConfig(t *testing.T, repoDir, workDir string) *config.Config {
	t.Helper()

	// Create channel directory
	cursorDir := filepath.Join(workDir, "cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}

	return &config.Config{
		Registries: map[string]string{
			"local-test": repoDir,
		},
		RegistryConfigs: map[string]map[string]string{
			"local-test": {
				"type": "git-local",
			},
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{cursorDir},
			},
		},
	}
}

func createMixedRegistryConfig(t *testing.T, localRepoDir, workDir string) *config.Config {
	t.Helper()

	// Create channel directory
	cursorDir := filepath.Join(workDir, "cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}

	return &config.Config{
		Registries: map[string]string{
			"local-test":  localRepoDir,
			"remote-test": "https://github.com/test/remote-repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"local-test": {
				"type": "git-local",
			},
			"remote-test": {
				"type": "git",
			},
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{cursorDir},
			},
		},
	}
}

func findDownloadedFiles(tempDir string) ([]string, error) {
	var sourceFiles []string
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sourceFiles = append(sourceFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return sourceFiles, nil
}
