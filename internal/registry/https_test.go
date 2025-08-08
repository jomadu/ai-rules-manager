package registry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPSRegistry(t *testing.T) {
	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     "https://example.com/registry",
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{
		Token: "test-token",
	}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if registry.GetName() != "test-https" {
		t.Errorf("Expected name 'test-https', got %s", registry.GetName())
	}

	if registry.GetType() != "https" {
		t.Errorf("Expected type 'https', got %s", registry.GetType())
	}
}

func TestHTTPSRegistry_GetRulesets(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			manifest := HTTPSManifest{
				Rulesets: map[string][]string{
					"python-rules": {"1.0.0", "1.1.0", "1.2.0"},
					"js-rules":     {"2.0.0", "2.1.0"},
				},
			}
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	rulesets, err := registry.GetRulesets(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rulesets) != 2 {
		t.Errorf("Expected 2 rulesets, got %d", len(rulesets))
	}

	// Check that we get the latest versions
	for _, ruleset := range rulesets {
		switch ruleset.Name {
		case "python-rules":
			if ruleset.Version != "1.2.0" {
				t.Errorf("Expected python-rules version 1.2.0, got %s", ruleset.Version)
			}
		case "js-rules":
			if ruleset.Version != "2.1.0" {
				t.Errorf("Expected js-rules version 2.1.0, got %s", ruleset.Version)
			}
		}
	}
}

func TestHTTPSRegistry_GetRuleset(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			manifest := HTTPSManifest{
				Rulesets: map[string][]string{
					"python-rules": {"1.0.0", "1.1.0", "1.2.0"},
				},
			}
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Test getting specific version
	ruleset, err := registry.GetRuleset(context.Background(), "python-rules", "1.1.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ruleset.Version != "1.1.0" {
		t.Errorf("Expected version 1.1.0, got %s", ruleset.Version)
	}

	// Test getting latest version
	ruleset, err = registry.GetRuleset(context.Background(), "python-rules", "latest")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ruleset.Version != "1.2.0" {
		t.Errorf("Expected latest version 1.2.0, got %s", ruleset.Version)
	}

	// Test non-existent ruleset
	_, err = registry.GetRuleset(context.Background(), "non-existent", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent ruleset")
	}

	// Test non-existent version
	_, err = registry.GetRuleset(context.Background(), "python-rules", "999.0.0")
	if err == nil {
		t.Error("Expected error for non-existent version")
	}
}

func TestHTTPSRegistry_DownloadRuleset(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/python-rules/1.0.0/ruleset.tar.gz" {
			// Check for Bearer token
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte("fake tar.gz content"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{
		Token: "test-token",
	}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "https-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test download
	err = registry.DownloadRuleset(context.Background(), "python-rules", "1.0.0", tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check file was created
	filePath := filepath.Join(tempDir, "ruleset.tar.gz")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected ruleset.tar.gz to be created")
	}

	// Check file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "fake tar.gz content" {
		t.Errorf("Expected 'fake tar.gz content', got %s", string(content))
	}
}

func TestHTTPSRegistry_GetVersions(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			manifest := HTTPSManifest{
				Rulesets: map[string][]string{
					"python-rules": {"1.0.0", "1.1.0", "1.2.0"},
					"empty-rules":  {},
				},
			}
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Test getting versions for existing ruleset
	versions, err := registry.GetVersions(context.Background(), "python-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := []string{"1.0.0", "1.1.0", "1.2.0"}
	if len(versions) != len(expected) {
		t.Errorf("Expected %d versions, got %d", len(expected), len(versions))
	}

	for i, version := range versions {
		if version != expected[i] {
			t.Errorf("Expected version %s, got %s", expected[i], version)
		}
	}

	// Test getting versions for ruleset with no versions
	versions, err = registry.GetVersions(context.Background(), "empty-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(versions) != 1 || versions[0] != "latest" {
		t.Errorf("Expected ['latest'] for empty versions, got %v", versions)
	}

	// Test non-existent ruleset
	_, err = registry.GetVersions(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent ruleset")
	}
}

func TestHTTPSRegistry_ManifestCaching(t *testing.T) {
	requestCount := 0

	// Create test server that counts requests
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			requestCount++
			manifest := HTTPSManifest{
				Rulesets: map[string][]string{
					"python-rules": {"1.0.0"},
				},
			}
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// First request should fetch manifest
	_, err = registry.GetRulesets(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Expected 1 request, got %d", requestCount)
	}

	// Second request should use cache
	_, err = registry.GetRulesets(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Expected 1 request (cached), got %d", requestCount)
	}
}

func TestHTTPSRegistry_InvalidManifest(t *testing.T) {
	// Create test server with invalid manifest
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			_, _ = w.Write([]byte(`{"invalid": "manifest"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &RegistryConfig{
		Name:    "test-https",
		Type:    "https",
		URL:     server.URL,
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{}

	registry, err := NewHTTPSRegistry(config, auth)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Skip TLS verification for test
	registry.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Should fail with invalid manifest
	_, err = registry.GetRulesets(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for invalid manifest")
	}

	if !strings.Contains(err.Error(), "missing required 'rulesets' field") {
		t.Errorf("Expected manifest validation error, got %v", err)
	}
}
