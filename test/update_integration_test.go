package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestUpdateScenario_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-update-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test Git repository with tags
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := createTestGitRepo(repoDir); err != nil {
		t.Fatal(err)
	}

	// Verify repository has expected tags
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		t.Fatal(err)
	}

	var tags []string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref.Name().Short())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedTags := []string{"1.0.0", "2.0.0"}
	if len(tags) != len(expectedTags) {
		t.Errorf("expected %d tags, got %d: %v", len(expectedTags), len(tags), tags)
	}

	for _, expected := range expectedTags {
		found := false
		for _, tag := range tags {
			if tag == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tag %s not found in %v", expected, tags)
		}
	}

	t.Logf("Successfully created test repository with tags: %v", tags)
}

func createTestGitRepo(repoDir string) error {
	// Initialize Git repository
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Create test file
	testFile := filepath.Join(repoDir, "rules", "test.md")
	if err := os.MkdirAll(filepath.Dir(testFile), 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(testFile, []byte("# Test Rule v1"), 0o644); err != nil {
		return err
	}

	_, err = worktree.Add("rules/test.md")
	if err != nil {
		return err
	}

	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		return err
	}

	// Create tag 1.0.0
	_, err = repo.CreateTag("1.0.0", commit1, nil)
	if err != nil {
		return err
	}

	// Update file for v2
	if err := os.WriteFile(testFile, []byte("# Test Rule v2"), 0o644); err != nil {
		return err
	}

	_, err = worktree.Add("rules/test.md")
	if err != nil {
		return err
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

	// Create tag 2.0.0
	_, err = repo.CreateTag("2.0.0", commit2, nil)
	return err
}
