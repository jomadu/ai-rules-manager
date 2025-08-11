package registry

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestGitRegistryVersionResolution_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "arm-git-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	repoDir := filepath.Join(tempDir, "test-repo")

	// Create a real Git repository with tags
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	// Add a test file
	testFile := filepath.Join(repoDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = worktree.Add("test.md")
	if err != nil {
		t.Fatal(err)
	}

	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create tag 1.0.0
	_, err = repo.CreateTag("1.0.0", commit1, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create another commit
	if err := os.WriteFile(testFile, []byte("# Test v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = worktree.Add("test.md")
	if err != nil {
		t.Fatal(err)
	}

	commit2, err := worktree.Commit("Second commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create tag 2.0.0
	_, err = repo.CreateTag("2.0.0", commit2, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create GitRegistry instance with minimal validation bypass
	config := &RegistryConfig{
		Name:    "test-registry",
		Type:    "git",
		URL:     "file://" + repoDir,
		Timeout: 30,
	}

	// Create registry directly without validation for testing
	gitReg := &GitRegistry{
		config: config,
		auth:   &AuthConfig{},
		client: &http.Client{},
	}

	ctx := context.Background()

	// Test version resolution
	tests := []struct {
		name        string
		versionSpec string
		expected    string
	}{
		{
			name:        "resolve >=1.0.0 to 2.0.0",
			versionSpec: ">=1.0.0",
			expected:    "2.0.0",
		},
		{
			name:        "resolve exact version 1.0.0",
			versionSpec: "1.0.0",
			expected:    "1.0.0",
		},
		{
			name:        "resolve ^1.0.0 to 1.0.0 (no 1.x.x > 1.0.0)",
			versionSpec: "^1.0.0",
			expected:    "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := gitReg.ResolveVersion(ctx, tt.versionSpec)
			if err != nil {
				t.Fatalf("ResolveVersion failed: %v", err)
			}

			if resolved != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resolved)
			}
		})
	}

	// Test GetVersions
	versions, err := gitReg.GetVersions(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	expectedVersions := []string{"latest", "1.0.0", "2.0.0"}
	if len(versions) != len(expectedVersions) {
		t.Errorf("expected %d versions, got %d", len(expectedVersions), len(versions))
	}

	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected version %s not found in %v", expected, versions)
		}
	}
}
