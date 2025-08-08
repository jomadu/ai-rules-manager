package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "no variables",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "$VAR format",
			input:    "$HOME/path",
			envVars:  map[string]string{"HOME": "/users/test"},
			expected: "/users/test/path",
		},
		{
			name:     "${VAR} format",
			input:    "${HOME}/path",
			envVars:  map[string]string{"HOME": "/users/test"},
			expected: "/users/test/path",
		},
		{
			name:     "missing variable",
			input:    "$MISSING/path",
			expected: "/path",
		},
		{
			name:     "multiple variables",
			input:    "$USER@$HOST",
			envVars:  map[string]string{"USER": "john", "HOST": "localhost"},
			expected: "john@localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadINIFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".armrc")

	// Create test INI file
	configContent := `[registries]
default = github.com/user/repo
my-s3 = my-bucket

[registries.default]
type = git
authToken = $GITHUB_TOKEN

[registries.my-s3]
type = s3
region = us-east-1

[git]
concurrency = 1
rateLimit = 10/minute

[network]
timeout = 30

[cache]
path = ~/.arm/cache
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set environment variable for testing
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Load configuration
	cfg := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		CacheConfig:     make(map[string]string),
	}

	err := cfg.loadINIFile(configPath, true)
	if err != nil {
		t.Fatalf("Failed to load INI file: %v", err)
	}

	// Test registries section
	if cfg.Registries["default"] != "github.com/user/repo" {
		t.Errorf("Expected default registry 'github.com/user/repo', got %q", cfg.Registries["default"])
	}

	// Test nested registry config with environment variable expansion
	if cfg.RegistryConfigs["default"]["authToken"] != "test-token" {
		t.Errorf("Expected authToken 'test-token', got %q", cfg.RegistryConfigs["default"]["authToken"])
	}

	// Test type defaults
	if cfg.TypeDefaults["git"]["concurrency"] != "1" {
		t.Errorf("Expected git concurrency '1', got %q", cfg.TypeDefaults["git"]["concurrency"])
	}

	// Test network config
	if cfg.NetworkConfig["timeout"] != "30" {
		t.Errorf("Expected network timeout '30', got %q", cfg.NetworkConfig["timeout"])
	}

	// Test cache config
	if cfg.CacheConfig["path"] != "~/.arm/cache" {
		t.Errorf("Expected cache path '~/.arm/cache', got %q", cfg.CacheConfig["path"])
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		CacheConfig:     make(map[string]string),
	}

	// Should not error for optional missing file
	err := cfg.loadINIFile("/nonexistent/file", false)
	if err != nil {
		t.Errorf("Expected no error for optional missing file, got: %v", err)
	}

	// Should error for required missing file
	err = cfg.loadINIFile("/nonexistent/file", true)
	if err == nil {
		t.Error("Expected error for required missing file")
	}
}
