package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewGitLabRegistry(t *testing.T) {
	config := &RegistryConfig{
		Name:    "test-gitlab",
		Type:    "gitlab",
		URL:     "https://gitlab.example.com/projects/123",
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{
		Token: "test-token",
	}

	registry, err := NewGitLabRegistry(config, auth)
	if err != nil {
		t.Fatalf("Failed to create GitLab registry: %v", err)
	}

	if registry.GetType() != "gitlab" {
		t.Errorf("Expected type 'gitlab', got %q", registry.GetType())
	}
	if registry.GetName() != "test-gitlab" {
		t.Errorf("Expected name 'test-gitlab', got %q", registry.GetName())
	}
	if registry.baseURL != "https://gitlab.example.com" {
		t.Errorf("Expected baseURL 'https://gitlab.example.com', got %q", registry.baseURL)
	}
	if registry.projectID != "123" {
		t.Errorf("Expected projectID '123', got %q", registry.projectID)
	}
}

func TestNewGitLabRegistryInvalidConfig(t *testing.T) {
	config := &RegistryConfig{
		Name: "test-gitlab",
		Type: "gitlab",
		// Missing URL
	}
	auth := &AuthConfig{}

	_, err := NewGitLabRegistry(config, auth)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestParseGitLabURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedBaseURL string
		expectedID      string
		expectError     bool
	}{
		{
			name:            "project URL",
			url:             "https://gitlab.example.com/projects/123",
			expectedBaseURL: "https://gitlab.example.com",
			expectedID:      "123",
			expectError:     false,
		},
		{
			name:            "group URL",
			url:             "https://gitlab.example.com/groups/456",
			expectedBaseURL: "https://gitlab.example.com",
			expectedID:      "456",
			expectError:     false,
		},
		{
			name:            "project URL with trailing slash",
			url:             "https://gitlab.example.com/projects/789/",
			expectedBaseURL: "https://gitlab.example.com",
			expectedID:      "789",
			expectError:     false,
		},
		{
			name:        "invalid URL format",
			url:         "https://github.com/user/repo",
			expectError: true,
		},
		{
			name:        "malformed URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "missing project ID",
			url:         "https://gitlab.example.com/projects/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, projectID, err := parseGitLabURL(tt.url)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if baseURL != tt.expectedBaseURL {
					t.Errorf("Expected baseURL %q, got %q", tt.expectedBaseURL, baseURL)
				}
				if projectID != tt.expectedID {
					t.Errorf("Expected projectID %q, got %q", tt.expectedID, projectID)
				}
			}
		})
	}
}

func TestGitLabGetRulesets(t *testing.T) {
	// Mock GitLab API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/packages") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": 1,
					"name": "python-rules",
					"version": "1.0.0",
					"package_type": "generic",
					"status": "default",
					"created_at": "2024-01-01T10:00:00Z",
					"updated_at": "2024-01-01T10:00:00Z"
				},
				{
					"id": 2,
					"name": "python-rules",
					"version": "1.1.0",
					"package_type": "generic",
					"status": "default",
					"created_at": "2024-01-02T10:00:00Z",
					"updated_at": "2024-01-02T10:00:00Z"
				},
				{
					"id": 3,
					"name": "javascript-rules",
					"version": "2.0.0",
					"package_type": "generic",
					"status": "default",
					"created_at": "2024-01-03T10:00:00Z",
					"updated_at": "2024-01-03T10:00:00Z"
				}
			]`))
		}
	}))
	defer server.Close()

	// Create registry with mock server
	registry := &GitLabRegistry{
		config: &RegistryConfig{
			Name: "test-gitlab",
			Type: "gitlab",
		},
		auth:      &AuthConfig{Token: "test-token"},
		client:    server.Client(),
		baseURL:   server.URL,
		projectID: "123",
	}

	ctx := context.Background()
	rulesets, err := registry.GetRulesets(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get rulesets: %v", err)
	}

	if len(rulesets) != 2 {
		t.Errorf("Expected 2 rulesets, got %d", len(rulesets))
	}

	// Check that we got the right rulesets with latest versions
	foundRulesets := make(map[string]string)
	for _, ruleset := range rulesets {
		foundRulesets[ruleset.Name] = ruleset.Version
		if ruleset.Type != "gitlab" {
			t.Errorf("Expected type 'gitlab', got %q", ruleset.Type)
		}
		if ruleset.Registry != "test-gitlab" {
			t.Errorf("Expected registry 'test-gitlab', got %q", ruleset.Registry)
		}
	}

	if foundRulesets["python-rules"] != "1.1.0" {
		t.Errorf("Expected python-rules version '1.1.0', got %q", foundRulesets["python-rules"])
	}
	if foundRulesets["javascript-rules"] != "2.0.0" {
		t.Errorf("Expected javascript-rules version '2.0.0', got %q", foundRulesets["javascript-rules"])
	}
}

func TestGitLabGetVersions(t *testing.T) {
	// Mock GitLab API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/packages") && strings.Contains(r.URL.RawQuery, "package_name=python-rules") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": 1,
					"name": "python-rules",
					"version": "1.0.0",
					"package_type": "generic"
				},
				{
					"id": 2,
					"name": "python-rules",
					"version": "1.1.0",
					"package_type": "generic"
				},
				{
					"id": 3,
					"name": "python-rules",
					"version": "1.2.0",
					"package_type": "generic"
				}
			]`))
		}
	}))
	defer server.Close()

	registry := &GitLabRegistry{
		config: &RegistryConfig{
			Name: "test-gitlab",
			Type: "gitlab",
		},
		auth:      &AuthConfig{Token: "test-token"},
		client:    server.Client(),
		baseURL:   server.URL,
		projectID: "123",
	}

	ctx := context.Background()
	versions, err := registry.GetVersions(ctx, "python-rules")
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	expectedVersions := []string{"1.0.0", "1.1.0", "1.2.0"}
	if len(versions) != len(expectedVersions) {
		t.Errorf("Expected %d versions, got %d", len(expectedVersions), len(versions))
	}

	for i, expected := range expectedVersions {
		if i < len(versions) && versions[i] != expected {
			t.Errorf("Expected version %q at index %d, got %q", expected, i, versions[i])
		}
	}
}

func TestGitLabRegistryClose(t *testing.T) {
	registry := &GitLabRegistry{
		config: &RegistryConfig{
			Name: "test-gitlab",
			Type: "gitlab",
		},
	}

	err := registry.Close()
	if err != nil {
		t.Errorf("Expected no error from Close(), got: %v", err)
	}
}
