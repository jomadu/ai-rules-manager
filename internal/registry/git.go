package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gobwas/glob"
	"github.com/jomadu/arm/pkg/types"
)

type GitRegistry struct {
	Name      string
	URL       string
	AuthToken string
	APIType   string // "github", "gitlab", "" (generic)

	// Clients will be initialized lazily
	cacheDir string
}

type GitMetadata struct {
	URL         string            `json:"url"`
	LastFetch   time.Time         `json:"lastFetch"`
	LastAccess  time.Time         `json:"lastAccess"`
	AccessCount int               `json:"accessCount"`
	Refs        map[string]string `json:"refs"` // ref name -> commit SHA
}

type GitReference struct {
	Type  ReferenceType
	Value string
}

type ReferenceType int

const (
	RefTypeBranch ReferenceType = iota
	RefTypeCommit
	RefTypeTag
	RefTypeDefault
)

func NewGitRegistry(name, url, authToken, apiType string) (*GitRegistry, error) {
	cacheDir, err := types.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	return &GitRegistry{
		Name:      name,
		URL:       url,
		AuthToken: authToken,
		APIType:   apiType,
		cacheDir:  filepath.Join(cacheDir, "git"),
	}, nil
}

// Registry interface implementation
func (r *GitRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	// For git registries, version is actually a git reference
	gitRef := r.parseReference(version)

	_, err := r.resolveReference(gitRef)
	if err != nil {
		return nil, err
	}

	return &types.Ruleset{
		Name:     name,
		Version:  version,
		Source:   r.URL,
		Files:    []string{}, // Will be populated during installation
		Checksum: "",         // Will be calculated during installation
	}, nil
}

func (r *GitRegistry) ListVersions(name string) ([]string, error) {
	metadata, err := r.getMetadata()
	if err != nil {
		return nil, err
	}

	var versions []string
	// Add branches
	for ref := range metadata.Refs {
		if strings.HasPrefix(ref, "refs/heads/") {
			branch := strings.TrimPrefix(ref, "refs/heads/")
			versions = append(versions, branch)
		}
	}
	// Add semver tags
	for ref := range metadata.Refs {
		if strings.HasPrefix(ref, "refs/tags/") {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			if isValidSemver(tag) {
				versions = append(versions, tag)
			}
		}
	}

	return versions, nil
}

func (r *GitRegistry) Download(name, version string) (io.ReadCloser, error) {
	// Git registries don't use traditional package downloads
	return nil, fmt.Errorf("git registries use file-based installation")
}

func (r *GitRegistry) GetMetadata(name string) (*Metadata, error) {
	gitMeta, err := r.getMetadata()
	if err != nil {
		return nil, err
	}

	versions, err := r.ListVersions(name)
	if err != nil {
		return nil, err
	}

	// Convert to Version structs
	var versionList []Version
	for _, v := range versions {
		versionList = append(versionList, Version{
			Version:   v,
			Published: gitMeta.LastFetch.Format(time.RFC3339),
		})
	}

	return &Metadata{
		Name:        name,
		Versions:    versionList,
		Description: fmt.Sprintf("Git repository: %s", r.URL),
		Repository:  r.URL,
	}, nil
}

func (r *GitRegistry) HealthCheck() error {
	_, err := r.getMetadata()
	return err
}

func (r *GitRegistry) InstallFiles(name, ref string, patterns []string, targetDir string) error {
	gitRef := r.parseReference(ref)

	// Get commit SHA for the reference
	commitSHA, err := r.resolveReference(gitRef)
	if err != nil {
		return err
	}

	// Get file list based on API type
	files, err := r.getFileList(commitSHA)
	if err != nil {
		return err
	}

	// Apply glob patterns
	matchedFiles := r.applyPatterns(files, patterns)
	if len(matchedFiles) == 0 {
		return fmt.Errorf("no files matched patterns: %v", patterns)
	}

	// Download and install matched files
	return r.downloadFiles(matchedFiles, commitSHA, targetDir)
}

func (r *GitRegistry) parseReference(ref string) *GitReference {
	if ref == "" {
		return &GitReference{Type: RefTypeDefault}
	}

	// Check if it's a commit SHA (40 or 7+ hex chars)
	if isCommitSHA(ref) {
		return &GitReference{Type: RefTypeCommit, Value: ref}
	}

	// Check if it's a semver tag
	if isValidSemver(ref) {
		return &GitReference{Type: RefTypeTag, Value: ref}
	}

	// Assume it's a branch
	return &GitReference{Type: RefTypeBranch, Value: ref}
}

func (r *GitRegistry) resolveReference(gitRef *GitReference) (string, error) {
	metadata, err := r.getMetadata()
	if err != nil {
		return "", err
	}

	switch gitRef.Type {
	case RefTypeDefault:
		// Use HEAD of default branch (usually main/master)
		if sha, ok := metadata.Refs["refs/heads/main"]; ok {
			return sha, nil
		}
		if sha, ok := metadata.Refs["refs/heads/master"]; ok {
			return sha, nil
		}
		return "", fmt.Errorf("no default branch found")

	case RefTypeBranch:
		refName := "refs/heads/" + gitRef.Value
		if sha, ok := metadata.Refs[refName]; ok {
			return sha, nil
		}
		return "", fmt.Errorf("branch not found: %s", gitRef.Value)

	case RefTypeTag:
		refName := "refs/tags/" + gitRef.Value
		if sha, ok := metadata.Refs[refName]; ok {
			return sha, nil
		}
		return "", fmt.Errorf("tag not found: %s", gitRef.Value)

	case RefTypeCommit:
		// For commits, we use the SHA directly but should validate it exists
		return gitRef.Value, nil

	default:
		return "", fmt.Errorf("unknown reference type")
	}
}

func (r *GitRegistry) getFileList(commitSHA string) ([]string, error) {
	switch r.APIType {
	case "github":
		return r.getGitHubFileList(commitSHA)
	case "gitlab":
		return r.getGitLabFileList(commitSHA)
	default:
		return r.getGenericGitFileList(commitSHA)
	}
}

func (r *GitRegistry) applyPatterns(files, patterns []string) []string {
	var matched []string

	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}

		for _, file := range files {
			if g.Match(file) {
				matched = append(matched, file)
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, file := range matched {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
		}
	}

	return unique
}

func (r *GitRegistry) getMetadata() (*GitMetadata, error) {
	repoPath := r.getRepoPath()
	metadataPath := filepath.Join(repoPath, "metadata.json")

	// Check if metadata exists and is recent
	if info, err := os.Stat(metadataPath); err == nil {
		if time.Since(info.ModTime()) < 5*time.Minute {
			return r.loadMetadata(metadataPath)
		}
	}

	// Fetch fresh metadata
	return r.fetchMetadata()
}

func (r *GitRegistry) getRepoPath() string {
	// Convert URL to cache path: https://github.com/owner/repo -> github.com/owner/repo
	url := strings.TrimPrefix(r.URL, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, ".git")
	return filepath.Join(r.cacheDir, url)
}

func (r *GitRegistry) loadMetadata(path string) (*GitMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata GitMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	// Update access time
	metadata.LastAccess = time.Now()
	metadata.AccessCount++

	// Save updated metadata
	_ = r.saveMetadata(&metadata)

	return &metadata, nil
}

func (r *GitRegistry) saveMetadata(metadata *GitMetadata) error {
	repoPath := r.getRepoPath()
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(repoPath, "metadata.json"), data, 0o644)
}

func isValidSemver(version string) bool {
	// Remove 'v' prefix if present
	v := strings.TrimPrefix(version, "v")
	_, err := semver.NewVersion(v)
	return err == nil
}

func isCommitSHA(ref string) bool {
	if len(ref) != 40 && len(ref) < 7 {
		return false
	}

	for _, c := range ref {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
