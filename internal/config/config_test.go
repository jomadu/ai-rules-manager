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

func TestLoadARMJSON(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "arm.json")

	// Create test JSON file
	jsonContent := `{
  "engines": {
    "arm": "^1.2.3"
  },
  "channels": {
    "cursor": {
      "directories": ["$HOME/.cursor/rules", "${PROJECT_ROOT}/custom"]
    },
    "q": {
      "directories": ["~/.aws/amazonq/rules"]
    }
  },
  "rulesets": {
    "default": {
      "my-rules": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "**/*.mdc"]
      },
      "python-rules": {
        "version": "~2.1.0"
      }
    },
    "my-registry": {
      "custom-rules": {
        "version": "latest"
      }
    }
  }
}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	// Set environment variables for testing
	os.Setenv("HOME", "/users/test")
	os.Setenv("PROJECT_ROOT", "/workspace/project")
	defer os.Unsetenv("HOME")
	defer os.Unsetenv("PROJECT_ROOT")

	// Load configuration
	cfg := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		CacheConfig:     make(map[string]string),
		Channels:        make(map[string]ChannelConfig),
		Rulesets:        make(map[string]map[string]RulesetSpec),
		Engines:         make(map[string]string),
	}

	err := cfg.loadARMJSON(jsonPath, true)
	if err != nil {
		t.Fatalf("Failed to load JSON file: %v", err)
	}

	// Test engines
	if cfg.Engines["arm"] != "^1.2.3" {
		t.Errorf("Expected arm engine '^1.2.3', got %q", cfg.Engines["arm"])
	}

	// Test channels with environment variable expansion
	if len(cfg.Channels["cursor"].Directories) != 2 {
		t.Errorf("Expected 2 cursor directories, got %d", len(cfg.Channels["cursor"].Directories))
	}
	if cfg.Channels["cursor"].Directories[0] != "/users/test/.cursor/rules" {
		t.Errorf("Expected '/users/test/.cursor/rules', got %q", cfg.Channels["cursor"].Directories[0])
	}
	if cfg.Channels["cursor"].Directories[1] != "/workspace/project/custom" {
		t.Errorf("Expected '/workspace/project/custom', got %q", cfg.Channels["cursor"].Directories[1])
	}

	// Test rulesets
	if cfg.Rulesets["default"]["my-rules"].Version != "^1.0.0" {
		t.Errorf("Expected version '^1.0.0', got %q", cfg.Rulesets["default"]["my-rules"].Version)
	}
	if len(cfg.Rulesets["default"]["my-rules"].Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(cfg.Rulesets["default"]["my-rules"].Patterns))
	}
}

func TestLoadLockFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "arm.lock")

	// Create test lock file
	lockContent := `{
  "rulesets": {
    "default": {
      "my-rules": {
        "version": "1.2.0",
        "resolved": "2024-01-15T10:30:00Z",
        "registry": "my-bucket",
        "type": "s3",
        "region": "us-east-1"
      }
    },
    "my-git": {
      "python-rules": {
        "version": "abc123def",
        "resolved": "2024-01-15T10:30:00Z",
        "registry": "https://github.com/user/repo",
        "type": "git"
      }
    }
  }
}`

	if err := os.WriteFile(lockPath, []byte(lockContent), 0600); err != nil {
		t.Fatalf("Failed to create test lock file: %v", err)
	}

	// Load configuration
	cfg := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		CacheConfig:     make(map[string]string),
		Channels:        make(map[string]ChannelConfig),
		Rulesets:        make(map[string]map[string]RulesetSpec),
		Engines:         make(map[string]string),
	}

	err := cfg.loadLockFile(lockPath)
	if err != nil {
		t.Fatalf("Failed to load lock file: %v", err)
	}

	// Test lock file content
	if cfg.LockFile == nil {
		t.Fatal("Expected lock file to be loaded")
	}

	lockedRuleset := cfg.LockFile.Rulesets["default"]["my-rules"]
	if lockedRuleset.Version != "1.2.0" {
		t.Errorf("Expected version '1.2.0', got %q", lockedRuleset.Version)
	}
	if lockedRuleset.Type != "s3" {
		t.Errorf("Expected type 's3', got %q", lockedRuleset.Type)
	}
	if lockedRuleset.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %q", lockedRuleset.Region)
	}
}

func TestExpandEnvVarsInJSON(t *testing.T) {
	jsonContent := `{
  "path": "$HOME/test",
  "url": "https://${HOST}:${PORT}/api",
  "missing": "$MISSING_VAR/path"
}`

	// Set environment variables
	os.Setenv("HOME", "/users/test")
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("HOME")
	defer os.Unsetenv("HOST")
	defer os.Unsetenv("PORT")

	expanded := expandEnvVarsInJSON(jsonContent)
	expected := `{
  "path": "/users/test/test",
  "url": "https://localhost:8080/api",
  "missing": "/path"
}`

	if expanded != expected {
		t.Errorf("Environment variable expansion failed\nGot: %s\nExpected: %s", expanded, expected)
	}
}
