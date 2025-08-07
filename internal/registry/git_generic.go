package registry

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Generic git implementations using git CLI
func (r *GitRegistry) getGenericGitFileList(commitSHA string) ([]string, error) {
	repoPath := r.getRepoPath()
	if err := r.ensureRepoCloned(repoPath); err != nil {
		return nil, err
	}
	return r.listFilesAtCommit(repoPath, commitSHA)
}

func (r *GitRegistry) downloadGenericGitFiles(files []string, commitSHA, targetDir string) error {
	repoPath := r.getRepoPath()
	if err := r.ensureRepoCloned(repoPath); err != nil {
		return err
	}
	return r.copyFilesFromRepo(repoPath, files, commitSHA, targetDir)
}

func (r *GitRegistry) fetchGenericGitMetadata() (*GitMetadata, error) {
	repoPath := r.getRepoPath()
	if err := r.ensureRepoCloned(repoPath); err != nil {
		return nil, err
	}
	refs := r.getRefsFromRepo(repoPath)
	return &GitMetadata{
		URL:         r.URL,
		LastFetch:   time.Now(),
		LastAccess:  time.Now(),
		AccessCount: 1,
		Refs:        refs,
	}, nil
}

func (r *GitRegistry) ensureRepoCloned(repoPath string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		return r.updateRepo(repoPath)
	}
	return r.cloneRepo(repoPath)
}

func (r *GitRegistry) cloneRepo(repoPath string) error {
	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", r.URL, repoPath)
	return cmd.Run()
}

func (r *GitRegistry) updateRepo(repoPath string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	return cmd.Run()
}

func (r *GitRegistry) listFilesAtCommit(repoPath, commitSHA string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", commitSHA)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}

func (r *GitRegistry) copyFilesFromRepo(repoPath string, files []string, commitSHA, targetDir string) error {
	for _, file := range files {
		cmd := exec.Command("git", "show", commitSHA+":"+file)
		cmd.Dir = repoPath
		content, err := cmd.Output()
		if err != nil {
			continue // Skip files that don't exist at this commit
		}
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

func (r *GitRegistry) getRefsFromRepo(repoPath string) map[string]string {
	refs := make(map[string]string)

	// Get branches
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads/")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) == 2 {
				refs[parts[0]] = parts[1]
			}
		}
	}

	// Get tags
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname) %(objectname)", "refs/tags/")
	cmd.Dir = repoPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) == 2 {
				refs[parts[0]] = parts[1]
			}
		}
	}

	return refs
}
