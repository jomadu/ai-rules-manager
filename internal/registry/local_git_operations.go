package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// LocalGitOperations implements GitOperations for local Git repositories
type LocalGitOperations struct {
	originalPath string // Store original path (may be symlink)
	repoPath     string // Resolved actual path
}

// NewLocalGitOperations creates a new local Git operations instance
func NewLocalGitOperations(repoPath string) (*LocalGitOperations, error) {
	// Store original path for configuration display
	originalPath := repoPath

	// Resolve symlinks and validate the repository path
	resolvedPath, err := resolveSymlinkPath(repoPath)
	if err != nil {
		return nil, err
	}

	return &LocalGitOperations{
		originalPath: originalPath,
		repoPath:     resolvedPath,
	}, nil
}

// ResolveVersion resolves a version spec to a concrete commit hash
func (l *LocalGitOperations) ResolveVersion(ctx context.Context, constraint string) (string, error) {
	if constraint == "latest" {
		return l.resolveDefaultBranch(ctx)
	}

	// Check if it's a commit hash (40 hex characters)
	if len(constraint) == 40 && IsHexString(constraint) {
		return constraint, nil
	}

	// Check if it's a semver pattern
	if IsSemverPattern(constraint) {
		versions, err := l.ListVersions(ctx)
		if err != nil {
			return "", l.enhanceGitError("list versions for semver resolution", err)
		}
		resolved, err := ResolveSemverPattern(constraint, versions)
		if err != nil {
			return "", l.enhanceGitError(fmt.Sprintf("resolve semver pattern %s", constraint), err)
		}
		return l.resolveTagToCommit(ctx, resolved)
	}

	// Check if it looks like a version number (tag)
	if IsVersionNumber(constraint) {
		return l.resolveTagToCommit(ctx, constraint)
	}

	// Assume it's a branch name and resolve to commit hash
	return l.resolveBranch(ctx, constraint)
}

// ListVersions returns available versions for the repository
func (l *LocalGitOperations) ListVersions(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "tag", "--list", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, l.enhanceGitError("tag --list --sort=-version:refname", err)
	}

	versions := []string{"latest"}
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, tag := range tags {
		if tag != "" {
			versions = append(versions, tag)
		}
	}

	return versions, nil
}

// GetFiles retrieves files matching patterns for a specific version
func (l *LocalGitOperations) GetFiles(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	// List all files at the specified version
	allFiles, err := l.listFilesAtVersion(ctx, version)
	if err != nil {
		return nil, l.enhanceGitError(fmt.Sprintf("list files at version %s", version), err)
	}

	// Apply pattern matching
	var matchingFiles []string
	for _, filePath := range allFiles {
		if MatchesAnyPattern(filePath, patterns) {
			matchingFiles = append(matchingFiles, filePath)
		}
	}

	if len(matchingFiles) == 0 {
		return nil, l.enhanceGitError(fmt.Sprintf("find files matching patterns %v at version %s", patterns, version), fmt.Errorf("no files match patterns"))
	}

	// Retrieve file contents using git show
	files := make(map[string][]byte)
	for _, filePath := range matchingFiles {
		content, err := l.getFileContent(ctx, version, filePath)
		if err != nil {
			return nil, l.enhanceGitError(fmt.Sprintf("get content for %s at version %s", filePath, version), err)
		}
		files[filePath] = content
	}

	return files, nil
}

// resolveDefaultBranch resolves "latest" to the default branch's HEAD commit
func (l *LocalGitOperations) resolveDefaultBranch(ctx context.Context) (string, error) {
	// Get the default branch name
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try common default branch names
		for _, branch := range []string{"main", "master"} {
			if hash, err := l.resolveBranch(ctx, branch); err == nil {
				return hash, nil
			}
		}
		return "", l.enhanceGitError("symbolic-ref refs/remotes/origin/HEAD", err)
	}

	// Extract branch name from refs/remotes/origin/main
	defaultRef := strings.TrimSpace(string(output))
	branchName := strings.TrimPrefix(defaultRef, "refs/remotes/origin/")
	return l.resolveBranch(ctx, branchName)
}

// resolveBranch resolves a branch name to its commit hash
func (l *LocalGitOperations) resolveBranch(ctx context.Context, branch string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "rev-parse", branch)
	output, err := cmd.Output()
	if err != nil {
		return "", l.enhanceGitError(fmt.Sprintf("rev-parse %s", branch), err)
	}

	return strings.TrimSpace(string(output)), nil
}

// resolveTagToCommit resolves a tag name to its commit hash, trying common formats
func (l *LocalGitOperations) resolveTagToCommit(ctx context.Context, tag string) (string, error) {
	// Try tag as-is first
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "rev-parse", tag)
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Try with 'v' prefix
	cmd = exec.CommandContext(ctx, "git", "-C", l.repoPath, "rev-parse", "v"+tag)
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "", l.enhanceGitError(fmt.Sprintf("resolve tag %s", tag), fmt.Errorf("tag not found: tried '%s' and 'v%s'", tag, tag))
}

// listFilesAtVersion lists all files in the repository at the specified version
func (l *LocalGitOperations) listFilesAtVersion(ctx context.Context, version string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "ls-tree", "-r", "--name-only", version)
	output, err := cmd.Output()
	if err != nil {
		return nil, l.enhanceGitError(fmt.Sprintf("ls-tree -r --name-only %s", version), err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

// getFileContent retrieves the content of a specific file at a version using git show
func (l *LocalGitOperations) getFileContent(ctx context.Context, version, filePath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", l.repoPath, "show", version+":"+filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, l.enhanceGitError(fmt.Sprintf("show %s:%s", version, filePath), err)
	}

	return output, nil
}

// resolveSymlinkPath resolves symlinks and validates the path is a Git repository
func resolveSymlinkPath(path string) (string, error) {
	// Use cross-platform path resolution
	absPath, err := config.ResolvePath(path)
	if err != nil {
		// Convert config errors to registry errors
		if strings.Contains(err.Error(), "permission denied") {
			return "", NewPermissionError(path, "resolve path")
		}
		if strings.Contains(err.Error(), "doesn't exist") {
			return "", NewRepositoryNotFoundError(path)
		}
		return "", NewPermissionError(path, "resolve path")
	}

	// Enhanced path validation with move detection
	if err := validateRepositoryPath(absPath); err != nil {
		return "", err
	}

	// Resolve symlinks
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", NewPermissionError(absPath, "resolve symlinks")
	}

	// Final validation on resolved path
	gitDir := filepath.Join(resolvedPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "", NewInvalidRepositoryError(resolvedPath)
	}

	return resolvedPath, nil
}

// validateRepositoryPath performs enhanced validation to detect moved repositories
func validateRepositoryPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check if parent directory exists to distinguish between moved vs completely missing
		parentDir := filepath.Dir(path)
		if _, parentErr := os.Stat(parentDir); parentErr == nil {
			// Parent exists but target doesn't - likely moved/renamed
			return NewRepositoryMovedError(path)
		}
		// Parent doesn't exist either - completely missing path
		return NewRepositoryNotFoundError(path)
	}

	// Path exists, check if it's a valid Git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return NewInvalidRepositoryError(path)
	}

	return nil
}

// enhanceGitError analyzes Git command errors and provides better context
func (l *LocalGitOperations) enhanceGitError(command string, err error) error {
	// Check if repository still exists
	if _, statErr := os.Stat(l.repoPath); os.IsNotExist(statErr) {
		return NewRepositoryMovedError(l.originalPath)
	}

	// Check if .git directory still exists
	gitDir := filepath.Join(l.repoPath, ".git")
	if _, statErr := os.Stat(gitDir); os.IsNotExist(statErr) {
		return NewCorruptedRepositoryError(l.repoPath, err)
	}

	// Return enhanced Git command error
	return NewGitCommandError(l.repoPath, command, err)
}

// GetOriginalPath returns the original path as stored in configuration (may be symlink)
func (l *LocalGitOperations) GetOriginalPath() string {
	return l.originalPath
}

// GetResolvedPath returns the actual resolved path used for Git operations
func (l *LocalGitOperations) GetResolvedPath() string {
	return l.repoPath
}
