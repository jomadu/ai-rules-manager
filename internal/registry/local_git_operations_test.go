package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// MockGitCommand helps mock Git command execution for testing
type MockGitCommand struct {
	commands map[string]string // command -> output mapping
	errors   map[string]error  // command -> error mapping
}

func NewMockGitCommand() *MockGitCommand {
	return &MockGitCommand{
		commands: make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *MockGitCommand) SetOutput(command, output string) {
	m.commands[command] = output
}

func (m *MockGitCommand) SetError(command string, err error) {
	m.errors[command] = err
}

func TestLocalGitOperations_ResolveVersion(t *testing.T) {
	// Create temporary Git repository for testing
	tempDir, cleanup := createTestGitRepo(t)
	defer cleanup()

	ops, err := NewLocalGitOperations(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		constraint  string
		expectError bool
		setup       func(t *testing.T)
	}{
		{
			name:       "resolve_latest",
			constraint: "latest",
			setup: func(t *testing.T) {
				// Create a commit to have a HEAD
				createTestCommit(t, tempDir, "Initial commit")
			},
		},
		{
			name:       "commit_hash",
			constraint: "1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:       "version_number",
			constraint: "1.0.0",
			setup: func(t *testing.T) {
				createTestCommit(t, tempDir, "Initial commit")
				createTestTag(t, tempDir, "1.0.0")
			},
		},
		{
			name:        "invalid_semver",
			constraint:  "^999.0.0",
			expectError: true,
			setup: func(t *testing.T) {
				createTestCommit(t, tempDir, "Initial commit")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := ops.ResolveVersion(ctx, tt.constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Errorf("Expected non-empty result")
			}
		})
	}
}

func TestLocalGitOperations_ListVersions(t *testing.T) {
	tempDir, cleanup := createTestGitRepo(t)
	defer cleanup()

	ops, err := NewLocalGitOperations(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations: %v", err)
	}

	// Create test tags
	createTestCommit(t, tempDir, "Initial commit")
	createTestTag(t, tempDir, "1.0.0")
	createTestTag(t, tempDir, "1.1.0")
	createTestTag(t, tempDir, "2.0.0")

	ctx := context.Background()
	versions, err := ops.ListVersions(ctx)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}

	// Should include "latest" plus the tags
	expectedCount := 4 // latest + 3 tags
	if len(versions) != expectedCount {
		t.Errorf("Expected %d versions, got %d: %v", expectedCount, len(versions), versions)
	}

	// Check that "latest" is included
	found := false
	for _, v := range versions {
		if v == "latest" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'latest' in versions list")
	}
}

func TestLocalGitOperations_GetFiles(t *testing.T) {
	tempDir, cleanup := createTestGitRepo(t)
	defer cleanup()

	ops, err := NewLocalGitOperations(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations: %v", err)
	}

	// Create test files and commit them
	testFiles := map[string]string{
		"rules/test1.md": "# Test 1",
		"rules/test2.md": "# Test 2",
		"docs/readme.md": "# README",
		"config.txt":     "config content",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	createTestCommit(t, tempDir, "Add test files")

	ctx := context.Background()

	tests := []struct {
		name         string
		patterns     []string
		expectedKeys []string
	}{
		{
			name:         "all_md_files",
			patterns:     []string{"*.md"},
			expectedKeys: []string{"rules/test1.md", "rules/test2.md", "docs/readme.md"},
		},
		{
			name:         "rules_only",
			patterns:     []string{"rules/*"},
			expectedKeys: []string{"rules/test1.md", "rules/test2.md"},
		},
		{
			name:         "specific_file",
			patterns:     []string{"config.txt"},
			expectedKeys: []string{"config.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := ops.GetFiles(ctx, "HEAD", tt.patterns)
			if err != nil {
				t.Fatalf("GetFiles failed: %v", err)
			}

			if len(files) != len(tt.expectedKeys) {
				t.Errorf("Expected %d files, got %d", len(tt.expectedKeys), len(files))
			}

			for _, expectedKey := range tt.expectedKeys {
				if _, exists := files[expectedKey]; !exists {
					t.Errorf("Expected file %s not found in results", expectedKey)
				}
			}
		})
	}
}

func TestLocalGitOperations_ErrorHandling(t *testing.T) {
	tempDir, cleanup := createTestGitRepo(t)
	defer cleanup()

	ops, err := NewLocalGitOperations(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations: %v", err)
	}

	ctx := context.Background()

	// Test with non-existent version
	_, err = ops.GetFiles(ctx, "nonexistent-version", []string{"*"})
	if err == nil {
		t.Errorf("Expected error for non-existent version")
	}

	// Test with invalid patterns that match no files
	createTestCommit(t, tempDir, "Initial commit")
	files, err := ops.GetFiles(ctx, "HEAD", []string{"*.nonexistent"})
	if err == nil {
		t.Errorf("Expected error for patterns matching no files")
	}
	if files != nil {
		t.Errorf("Expected nil files for error case")
	}
}

// Helper functions for creating test Git repositories

func createTestGitRepo(t *testing.T) (repoPath string, cleanup func()) {
	tempDir, err := os.MkdirTemp("", "git-ops-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize Git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tempDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure Git user for commits
	configCmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmdArgs := range configCmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			_ = os.RemoveAll(tempDir)
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	cleanup = func() {
		_ = os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func createTestCommit(t *testing.T, repoDir, message string) {
	// Create a unique dummy file for each commit
	dummyFile := filepath.Join(repoDir, fmt.Sprintf("dummy-%d.txt", time.Now().UnixNano()))
	if err := os.WriteFile(dummyFile, []byte(fmt.Sprintf("dummy content %s", message)), 0o644); err != nil {
		t.Fatal(err)
	}

	// Add and commit
	cmds := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", message},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run %v: %v", cmdArgs, err)
		}
	}
}

func createTestTag(t *testing.T, repoDir, tagName string) {
	cmd := exec.Command("git", "tag", tagName)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create tag %s: %v", tagName, err)
	}
}
