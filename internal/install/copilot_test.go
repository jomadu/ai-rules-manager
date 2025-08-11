package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestInstallCopilotChannel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create test source files
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create copilot test files
	copilotFiles := map[string]string{
		"copilot-instructions.md": "# Test Instructions\nTest content for copilot instructions.",
		"copilot-prompts.yml":     "prompts:\n  - name: test\n    description: Test prompt\n    content: Test content",
		"copilot-chat-participants.yml": "participants:\n  - name: test\n    description: Test participant",
	}

	var sourceFiles []string
	for filename, content := range copilotFiles {
		filePath := filepath.Join(sourceDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		sourceFiles = append(sourceFiles, filePath)
	}

	// Create target .github directory
	githubDir := filepath.Join(tempDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("Failed to create .github directory: %v", err)
	}

	// Create config with copilot channel
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"copilot": {
				Directories: []string{githubDir},
			},
		},
	}

	// Create installer
	installer := New(cfg)

	// Test installation request
	req := &InstallRequest{
		Registry:    "test-registry",
		Ruleset:     "copilot-rules",
		Version:     "1.0.0",
		SourceFiles: sourceFiles,
		Channels:    []string{"copilot"},
	}

	// Perform installation
	result, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Installation failed: %v", err)
	}

	// Verify result
	if result.Registry != "test-registry" {
		t.Errorf("Expected registry 'test-registry', got '%s'", result.Registry)
	}
	if result.Ruleset != "copilot-rules" {
		t.Errorf("Expected ruleset 'copilot-rules', got '%s'", result.Ruleset)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.Version)
	}
	if len(result.Channels) != 1 || result.Channels[0] != "copilot" {
		t.Errorf("Expected channels ['copilot'], got %v", result.Channels)
	}
	if result.FilesCount != len(copilotFiles) {
		t.Errorf("Expected %d files, got %d", len(copilotFiles), result.FilesCount)
	}

	// Verify files were copied to .github directory
	for filename := range copilotFiles {
		targetPath := filepath.Join(githubDir, filename)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("File %s was not copied to .github directory", filename)
		}
	}

	// Verify file contents
	for filename, expectedContent := range copilotFiles {
		targetPath := filepath.Join(githubDir, filename)
		actualContent, err := os.ReadFile(targetPath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filename, err)
			continue
		}
		if string(actualContent) != expectedContent {
			t.Errorf("File %s content mismatch.\nExpected: %s\nActual: %s", filename, expectedContent, string(actualContent))
		}
	}
}

func TestCopilotChannelWithMultipleDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create test source file
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	sourceFile := filepath.Join(sourceDir, "copilot-instructions.md")
	testContent := "# Test Instructions\nTest content for copilot instructions."
	if err := os.WriteFile(sourceFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create multiple target directories
	githubDir1 := filepath.Join(tempDir, "project1", ".github")
	githubDir2 := filepath.Join(tempDir, "project2", ".github")
	
	for _, dir := range []string{githubDir1, githubDir2} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create config with copilot channel pointing to multiple directories
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"copilot": {
				Directories: []string{githubDir1, githubDir2},
			},
		},
	}

	// Create installer
	installer := New(cfg)

	// Test installation request
	req := &InstallRequest{
		Registry:    "test-registry",
		Ruleset:     "copilot-rules",
		Version:     "1.0.0",
		SourceFiles: []string{sourceFile},
		Channels:    []string{"copilot"},
	}

	// Perform installation
	result, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Installation failed: %v", err)
	}

	// Verify files were copied to both directories
	for _, dir := range []string{githubDir1, githubDir2} {
		targetPath := filepath.Join(dir, "copilot-instructions.md")
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("File was not copied to directory %s", dir)
		}

		// Verify content
		actualContent, err := os.ReadFile(targetPath)
		if err != nil {
			t.Errorf("Failed to read file in %s: %v", dir, err)
			continue
		}
		if string(actualContent) != testContent {
			t.Errorf("File content mismatch in %s", dir)
		}
	}

	// Verify the result shows the correct number of files (should be 2 since we copy to 2 directories)
	if result.FilesCount != 2 {
		t.Errorf("Expected 2 files (1 file copied to 2 directories), got %d", result.FilesCount)
	}
}