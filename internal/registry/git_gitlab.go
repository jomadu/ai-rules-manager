package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type GitLabTreeItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

func (r *GitRegistry) getGitLabFileList(commitSHA string) ([]string, error) {
	projectID, baseURL, err := r.parseGitLabURL()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?recursive=true&ref=%s", baseURL, projectID, commitSHA)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("PRIVATE-TOKEN", r.AuthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	var items []GitLabTreeItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var files []string
	for _, item := range items {
		if item.Type == "blob" { // Only include files, not directories
			files = append(files, item.Path)
		}
	}

	return files, nil
}

func (r *GitRegistry) downloadGitLabFiles(files []string, commitSHA, targetDir string) error {
	projectID, baseURL, err := r.parseGitLabURL()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.downloadGitLabFile(projectID, baseURL, file, commitSHA, targetDir); err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
	}

	return nil
}

func (r *GitRegistry) downloadGitLabFile(projectID, baseURL, filePath, commitSHA, targetDir string) error {
	// URL encode the file path
	encodedPath := url.QueryEscape(filePath)
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s", baseURL, projectID, encodedPath, commitSHA)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err
	}

	if r.AuthToken != "" {
		req.Header.Set("PRIVATE-TOKEN", r.AuthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	// Read file content directly (GitLab returns raw content)
	content, err := io.ReadAll(resp.Body)
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

func (r *GitRegistry) fetchGitLabMetadata() (*GitMetadata, error) {
	projectID, baseURL, err := r.parseGitLabURL()
	if err != nil {
		return nil, err
	}

	refs := make(map[string]string)

	// Get branches
	branches, err := r.getGitLabBranches(projectID, baseURL)
	if err != nil {
		return nil, err
	}
	for name, sha := range branches {
		refs["refs/heads/"+name] = sha
	}

	// Get tags
	tags, err := r.getGitLabTags(projectID, baseURL)
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

func (r *GitRegistry) getGitLabBranches(projectID, baseURL string) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches", baseURL, projectID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("PRIVATE-TOKEN", r.AuthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	var branches []struct {
		Name   string `json:"name"`
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, branch := range branches {
		result[branch.Name] = branch.Commit.ID
	}

	return result, nil
}

func (r *GitRegistry) getGitLabTags(projectID, baseURL string) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/tags", baseURL, projectID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if r.AuthToken != "" {
		req.Header.Set("PRIVATE-TOKEN", r.AuthToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, tag := range tags {
		result[tag.Name] = tag.Commit.ID
	}

	return result, nil
}

func (r *GitRegistry) parseGitLabURL() (projectID, baseURL string, err error) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return "", "", err
	}

	// Extract base URL (scheme + host)
	baseURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	// For GitLab, we need to determine the project ID
	// This is a simplified approach - in practice, you might need to:
	// 1. Use the project path to look up the ID via API
	// 2. Store the project ID in configuration
	// 3. Parse it from the URL if it's numeric

	path := strings.Trim(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	// Try to parse as numeric project ID first
	if id, err := strconv.Atoi(path); err == nil {
		return strconv.Itoa(id), baseURL, nil
	}

	// If not numeric, URL encode the path as project ID
	// GitLab API accepts both numeric IDs and URL-encoded paths
	projectID = url.QueryEscape(path)

	return projectID, baseURL, nil
}
