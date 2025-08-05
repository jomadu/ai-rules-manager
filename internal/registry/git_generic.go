package registry

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (r *GitRegistry) getGenericGitFileList(commitSHA string) ([]string, error) {
	repoPath := r.getRepoPath()
	gitDir := filepath.Join(repoPath, ".git")

	// Ensure repository is cloned and up to date
	if err := r.ensureRepository(); err != nil {
		return nil, err
	}

	// Use git ls-tree to list files at commit
	cmd := exec.Command("git", "--git-dir", gitDir, "ls-tree", "-r", "--name-only", commitSHA)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

func (r *GitRegistry) downloadGenericGitFiles(files []string, commitSHA, targetDir string) error {
	repoPath := r.getRepoPath()
	gitDir := filepath.Join(repoPath, ".git")

	for _, file := range files {
		// Use git show to get file content
		cmd := exec.Command("git", "--git-dir", gitDir, "show", commitSHA+":"+file)
		content, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get file %s: %w", file, err)
		}

		// Write file to target directory
		targetPath := filepath.Join(targetDir, file)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitRegistry) fetchGenericGitMetadata() (*GitMetadata, error) {
	if err := r.ensureRepository(); err != nil {
		return nil, err
	}

	repoPath := r.getRepoPath()
	gitDir := filepath.Join(repoPath, ".git")

	// Get all refs
	cmd := exec.Command("git", "--git-dir", gitDir, "for-each-ref", "--format=%(refname) %(objectname)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref failed: %w", err)
	}

	refs := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			refs[parts[0]] = parts[1]
		}
	}

	metadata := &GitMetadata{
		URL:         r.URL,
		LastFetch:   time.Now(),
		LastAccess:  time.Now(),
		AccessCount: 1,
		Refs:        refs,
	}

	if err := r.saveMetadata(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (r *GitRegistry) ensureRepository() error {
	repoPath := r.getRepoPath()
	gitDir := filepath.Join(repoPath, ".git")

	// Check if repository exists
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone as bare repository
		if err := os.MkdirAll(repoPath, 0o755); err != nil {
			return err
		}

		cmd := exec.Command("git", "clone", "--bare", r.getAuthURL(), gitDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		// Fetch updates
		cmd := exec.Command("git", "--git-dir", gitDir, "fetch", "--all")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git fetch failed: %w", err)
		}
	}

	return nil
}

func (r *GitRegistry) getAuthURL() string {
	if r.AuthToken == "" {
		return r.URL
	}

	// Add token to URL for HTTPS authentication
	if strings.HasPrefix(r.URL, "https://") {
		// For GitHub: https://token@github.com/owner/repo.git
		// For GitLab: https://oauth2:token@gitlab.com/owner/repo.git
		url := strings.TrimPrefix(r.URL, "https://")

		if strings.Contains(r.URL, "github.com") {
			return fmt.Sprintf("https://%s@%s", r.AuthToken, url)
		} else {
			// Assume GitLab-style for other providers
			return fmt.Sprintf("https://oauth2:%s@%s", r.AuthToken, url)
		}
	}

	return r.URL
}

// Helper function to decode base64 content (used by GitHub API)
func decodeBase64(encoded string) ([]byte, error) {
	// Remove whitespace and newlines
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, " ", "")

	return base64.StdEncoding.DecodeString(encoded)
}
