package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://registry.armjs.org/", "registry.armjs.org"},
		{"https://gitlab.company.com/api/v4", "gitlab.company.com"},
		{"http://localhost:8080", "localhost:8080"},
		{"file:///tmp/registry", "_tmp_registry"},
		{"", "unknown"},
		{"invalid-url", "hash-"},
	}

	for _, test := range tests {
		result := extractHost(test.url)
		if test.expected == "hash-" {
			if len(result) < 13 || result[:5] != "hash-" {
				t.Errorf("extractHost(%q) = %q, expected hash prefix", test.url, result)
			}
		} else if result != test.expected {
			t.Errorf("extractHost(%q) = %q, expected %q", test.url, result, test.expected)
		}
	}
}

func TestStoragePaths(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}

	packagePath := storage.PackagePath("https://registry.armjs.org/", "typescript-rules", "1.0.0")
	expected := filepath.Join(tmpDir, "packages", "registry.armjs.org", "typescript-rules", "1.0.0")
	if packagePath != expected {
		t.Errorf("PackagePath() = %q, expected %q", packagePath, expected)
	}

	metadataPath := storage.MetadataPath("https://registry.armjs.org/")
	expected = filepath.Join(tmpDir, "registry", "registry.armjs.org")
	if metadataPath != expected {
		t.Errorf("MetadataPath() = %q, expected %q", metadataPath, expected)
	}
}

func TestStorageExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test never expires (TTL = 0)
	if storage.IsExpired(testFile, 0) {
		t.Error("File should never expire with TTL = 0")
	}

	// Test not expired
	if storage.IsExpired(testFile, time.Hour) {
		t.Error("File should not be expired")
	}

	// Test expired (set file time to past)
	pastTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(testFile, pastTime, pastTime); err != nil {
		t.Fatal(err)
	}

	if !storage.IsExpired(testFile, time.Hour) {
		t.Error("File should be expired")
	}

	// Test non-existent file
	if !storage.IsExpired("nonexistent", time.Hour) {
		t.Error("Non-existent file should be considered expired")
	}
}
