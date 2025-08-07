package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GitHubTreeResponse struct {
	Tree []GitHubTreeItem `json:"tree"`
}

type GitHubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type GitHubContentsResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (r *GitRegistry) getGitHubFileList(commitSHA string) ([]string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, commitSHA)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("Authorization", "token "+r.AuthToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var treeResp GitHubTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, err
	}

	var files []string
	for _, item := range treeResp.Tree {
		if item.Type == "blob" { // Only include files, not directories
			files = append(files, item.Path)
		}
	}

	return files, nil
}

func (r *GitRegistry) downloadFiles(files []string, commitSHA, targetDir string) error {
	switch r.APIType {
	case "github":
		err := r.downloadGitHubFiles(files, commitSHA, targetDir)
		if err == nil {
			return nil
		}
		// Fall back to generic git on API failure
		return r.downloadGenericGitFiles(files, commitSHA, targetDir)
	case "gitlab":
		err := r.downloadGitLabFiles(files, commitSHA, targetDir)
		if err == nil {
			return nil
		}
		// Fall back to generic git on API failure
		return r.downloadGenericGitFiles(files, commitSHA, targetDir)
	default:
		return r.downloadGenericGitFiles(files, commitSHA, targetDir)
	}
}

func (r *GitRegistry) downloadGitHubFiles(files []string, commitSHA, targetDir string) error {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.downloadGitHubFile(owner, repo, file, commitSHA, targetDir); err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
	}

	return nil
}

func (r *GitRegistry) downloadGitHubFile(owner, repo, filePath, commitSHA, targetDir string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, filePath, commitSHA)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err
	}

	if r.AuthToken != "" {
		req.Header.Set("Authorization", "token "+r.AuthToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var contents GitHubContentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return err
	}

	// GitHub returns base64 encoded content
	if contents.Encoding != "base64" {
		return fmt.Errorf("unexpected encoding: %s", contents.Encoding)
	}

	// Decode base64 content
	content, err := decodeBase64(contents.Content)
	if err != nil {
		return err
	}

	// Write file to target directory
	targetPath := filepath.Join(targetDir, filePath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(targetPath, content, 0o644)
}

func (r *GitRegistry) fetchMetadata() (*GitMetadata, error) {
	switch r.APIType {
	case "github":
		meta, err := r.fetchGitHubMetadata()
		if err == nil {
			return meta, nil
		}
		// Fall back to generic git on API failure
		return r.fetchGenericGitMetadata()
	case "gitlab":
		meta, err := r.fetchGitLabMetadata()
		if err == nil {
			return meta, nil
		}
		// Fall back to generic git on API failure
		return r.fetchGenericGitMetadata()
	default:
		return r.fetchGenericGitMetadata()
	}
}

func (r *GitRegistry) fetchGitHubMetadata() (*GitMetadata, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return nil, err
	}

	refs := make(map[string]string)

	// Get branches
	branches, err := r.getGitHubBranches(owner, repo)
	if err != nil {
		return nil, err
	}
	for name, sha := range branches {
		refs["refs/heads/"+name] = sha
	}

	// Get tags
	tags, err := r.getGitHubTags(owner, repo)
	if err != nil {
		return nil, err
	}
	for name, sha := range tags {
		refs["refs/tags/"+name] = sha
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

func (r *GitRegistry) getGitHubBranches(owner, repo string) (map[string]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("Authorization", "token "+r.AuthToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var branches []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, branch := range branches {
		result[branch.Name] = branch.Commit.SHA
	}

	return result, nil
}

func (r *GitRegistry) getGitHubTags(owner, repo string) (map[string]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("Authorization", "token "+r.AuthToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, tag := range tags {
		result[tag.Name] = tag.Commit.SHA
	}

	return result, nil
}

func (r *GitRegistry) parseGitHubURL() (owner, repo string, err error) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return "", "", err
	}

	if u.Host != "github.com" {
		return "", "", fmt.Errorf("not a GitHub URL: %s", r.URL)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", r.URL)
	}

	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")

	return owner, repo, nil
}

// decodeBase64 decodes base64 content from GitHub API
func decodeBase64(content string) ([]byte, error) {
	// Remove whitespace and newlines
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")

	return base64.StdEncoding.DecodeString(content)
}
