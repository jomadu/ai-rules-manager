package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
			var varsToUnset []string
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
				varsToUnset = append(varsToUnset, k)
			}
			defer func() {
				for _, k := range varsToUnset {
					_ = os.Unsetenv(k)
				}
			}()

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

	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set environment variable for testing
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	defer func() { _ = os.Unsetenv("GITHUB_TOKEN") }()

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

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0o600); err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	// Set environment variables for testing
	_ = os.Setenv("HOME", "/users/test")
	_ = os.Setenv("PROJECT_ROOT", "/workspace/project")
	defer func() { _ = os.Unsetenv("HOME") }()
	defer func() { _ = os.Unsetenv("PROJECT_ROOT") }()

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

	if err := os.WriteFile(lockPath, []byte(lockContent), 0o600); err != nil {
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
	_ = os.Setenv("HOME", "/users/test")
	_ = os.Setenv("HOST", "localhost")
	_ = os.Setenv("PORT", "8080")
	defer func() { _ = os.Unsetenv("HOME") }()
	defer func() { _ = os.Unsetenv("HOST") }()
	defer func() { _ = os.Unsetenv("PORT") }()

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

func TestMergeConfigs(t *testing.T) {
	// Create global config
	global := &Config{
		Registries: map[string]string{
			"default": "github.com/global/repo",
			"shared":  "shared-registry",
		},
		RegistryConfigs: map[string]map[string]string{
			"default": {
				"type":      "git",
				"authToken": "global-token",
			},
			"shared": {
				"type":        "s3",
				"region":      "us-west-1",
				"concurrency": "5",
			},
		},
		TypeDefaults: map[string]map[string]string{
			"git": {
				"concurrency": "1",
				"rateLimit":   "10/minute",
			},
		},
		NetworkConfig: map[string]string{
			"timeout": "30",
			"retries": "3",
		},
		Engines: map[string]string{
			"arm": "^1.0.0",
		},
		Channels: map[string]ChannelConfig{
			"cursor": {
				Directories: []string{"/global/cursor"},
			},
		},
		Rulesets: map[string]map[string]RulesetSpec{
			"default": {
				"global-rules": {Version: "1.0.0"},
				"shared-rules": {Version: "1.0.0"},
			},
		},
	}

	// Create local config
	local := &Config{
		Registries: map[string]string{
			"default": "github.com/local/repo", // Override global
			"local":   "local-registry",        // New key
		},
		RegistryConfigs: map[string]map[string]string{
			"default": {
				"authToken": "local-token", // Override global
				"apiType":   "github",      // New key
			},
			"local": {
				"type": "local", // New registry config
			},
		},
		TypeDefaults: map[string]map[string]string{
			"git": {
				"concurrency": "2", // Override global
			},
			"s3": {
				"region": "us-east-1", // New type
			},
		},
		NetworkConfig: map[string]string{
			"timeout": "60", // Override global
		},
		Engines: map[string]string{
			"arm": "^1.2.0", // Override global
		},
		Channels: map[string]ChannelConfig{
			"cursor": {
				Directories: []string{"/local/cursor"}, // Override global
			},
			"q": {
				Directories: []string{"/local/q"}, // New channel
			},
		},
		Rulesets: map[string]map[string]RulesetSpec{
			"default": {
				"shared-rules": {Version: "2.0.0"}, // Override global
				"local-rules":  {Version: "1.0.0"}, // New ruleset
			},
			"local": {
				"local-only": {Version: "1.0.0"}, // New registry
			},
		},
	}

	// Merge configurations
	merged := mergeConfigs(global, local)

	// Test registries merge
	if merged.Registries["default"] != "github.com/local/repo" {
		t.Errorf("Expected local override for default registry, got %q", merged.Registries["default"])
	}
	if merged.Registries["shared"] != "shared-registry" {
		t.Errorf("Expected global value for shared registry, got %q", merged.Registries["shared"])
	}
	if merged.Registries["local"] != "local-registry" {
		t.Errorf("Expected local value for local registry, got %q", merged.Registries["local"])
	}

	// Test nested registry configs merge
	if merged.RegistryConfigs["default"]["type"] != "git" {
		t.Errorf("Expected global type to be preserved, got %q", merged.RegistryConfigs["default"]["type"])
	}
	if merged.RegistryConfigs["default"]["authToken"] != "local-token" {
		t.Errorf("Expected local override for authToken, got %q", merged.RegistryConfigs["default"]["authToken"])
	}
	if merged.RegistryConfigs["default"]["apiType"] != "github" {
		t.Errorf("Expected local apiType to be added, got %q", merged.RegistryConfigs["default"]["apiType"])
	}
	if merged.RegistryConfigs["shared"]["concurrency"] != "5" {
		t.Errorf("Expected global shared config to be preserved, got %q", merged.RegistryConfigs["shared"]["concurrency"])
	}

	// Test type defaults merge
	if merged.TypeDefaults["git"]["concurrency"] != "2" {
		t.Errorf("Expected local override for git concurrency, got %q", merged.TypeDefaults["git"]["concurrency"])
	}
	if merged.TypeDefaults["git"]["rateLimit"] != "10/minute" {
		t.Errorf("Expected global rateLimit to be preserved, got %q", merged.TypeDefaults["git"]["rateLimit"])
	}
	if merged.TypeDefaults["s3"]["region"] != "us-east-1" {
		t.Errorf("Expected local s3 config to be added, got %q", merged.TypeDefaults["s3"]["region"])
	}

	// Test network config merge
	if merged.NetworkConfig["timeout"] != "60" {
		t.Errorf("Expected local override for timeout, got %q", merged.NetworkConfig["timeout"])
	}
	if merged.NetworkConfig["retries"] != "3" {
		t.Errorf("Expected global retries to be preserved, got %q", merged.NetworkConfig["retries"])
	}

	// Test engines merge
	if merged.Engines["arm"] != "^1.2.0" {
		t.Errorf("Expected local override for arm version, got %q", merged.Engines["arm"])
	}

	// Test channels merge
	if len(merged.Channels["cursor"].Directories) != 1 || merged.Channels["cursor"].Directories[0] != "/local/cursor" {
		t.Errorf("Expected local override for cursor channel, got %v", merged.Channels["cursor"].Directories)
	}
	if len(merged.Channels["q"].Directories) != 1 || merged.Channels["q"].Directories[0] != "/local/q" {
		t.Errorf("Expected local q channel to be added, got %v", merged.Channels["q"].Directories)
	}

	// Test rulesets merge
	if merged.Rulesets["default"]["global-rules"].Version != "1.0.0" {
		t.Errorf("Expected global-rules to be preserved, got %q", merged.Rulesets["default"]["global-rules"].Version)
	}
	if merged.Rulesets["default"]["shared-rules"].Version != "2.0.0" {
		t.Errorf("Expected local override for shared-rules, got %q", merged.Rulesets["default"]["shared-rules"].Version)
	}
	if merged.Rulesets["default"]["local-rules"].Version != "1.0.0" {
		t.Errorf("Expected local-rules to be added, got %q", merged.Rulesets["default"]["local-rules"].Version)
	}
	if merged.Rulesets["local"]["local-only"].Version != "1.0.0" {
		t.Errorf("Expected local registry to be added, got %q", merged.Rulesets["local"]["local-only"].Version)
	}
}

func TestHierarchicalLoad(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	globalDir := filepath.Join(tmpDir, ".arm")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("Failed to create global dir: %v", err)
	}

	// Create global .armrc
	globalINI := `[registries]
default = https://github.com/global/repo
shared = shared-registry

[registries.default]
type = git
authToken = global-token

[registries.shared]
type = s3
region = us-west-1

[git]
concurrency = 1
rateLimit = 10/minute`

	if err := os.WriteFile(filepath.Join(globalDir, ".armrc"), []byte(globalINI), 0o600); err != nil {
		t.Fatalf("Failed to create global .armrc: %v", err)
	}

	// Create global arm.json
	globalJSON := `{
  "engines": {"arm": "^1.0.0"},
  "channels": {"cursor": {"directories": ["/global/cursor"]}},
  "rulesets": {"default": {"global-rules": {"version": "1.0.0"}}}
}`

	if err := os.WriteFile(filepath.Join(globalDir, "arm.json"), []byte(globalJSON), 0o600); err != nil {
		t.Fatalf("Failed to create global arm.json: %v", err)
	}

	// Create local .armrc
	localINI := `[registries]
default = https://github.com/local/repo
local = /path/to/local

[registries.default]
authToken = local-token
apiType = github

[registries.local]
type = local

[git]
concurrency = 2`

	if err := os.WriteFile(filepath.Join(tmpDir, ".armrc"), []byte(localINI), 0o600); err != nil {
		t.Fatalf("Failed to create local .armrc: %v", err)
	}

	// Create local arm.json
	localJSON := `{
  "engines": {"arm": "^1.2.0"},
  "channels": {"cursor": {"directories": ["/local/cursor"]}, "q": {"directories": ["/local/q"]}},
  "rulesets": {"default": {"local-rules": {"version": "1.0.0"}}}
}`

	if err := os.WriteFile(filepath.Join(tmpDir, "arm.json"), []byte(localJSON), 0o600); err != nil {
		t.Fatalf("Failed to create local arm.json: %v", err)
	}

	// Set HOME to tmpDir for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Change to tmpDir for local file loading
	originalWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(originalWd) }()

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load hierarchical config: %v", err)
	}

	// Test merged results
	if cfg.Registries["default"] != "https://github.com/local/repo" {
		t.Errorf("Expected local override for default registry, got %q", cfg.Registries["default"])
	}
	if cfg.Registries["shared"] != "shared-registry" {
		t.Errorf("Expected global shared registry, got %q", cfg.Registries["shared"])
	}
	if cfg.RegistryConfigs["default"]["type"] != "git" {
		t.Errorf("Expected global type preserved, got %q", cfg.RegistryConfigs["default"]["type"])
	}
	if cfg.RegistryConfigs["default"]["authToken"] != "local-token" {
		t.Errorf("Expected local authToken override, got %q", cfg.RegistryConfigs["default"]["authToken"])
	}
	if cfg.TypeDefaults["git"]["concurrency"] != "2" {
		t.Errorf("Expected local git concurrency override, got %q", cfg.TypeDefaults["git"]["concurrency"])
	}
	if cfg.TypeDefaults["git"]["rateLimit"] != "10/minute" {
		t.Errorf("Expected global git rateLimit preserved, got %q", cfg.TypeDefaults["git"]["rateLimit"])
	}
	if cfg.Engines["arm"] != "^1.2.0" {
		t.Errorf("Expected local arm version override, got %q", cfg.Engines["arm"])
	}
}

func TestValidateRegistry(t *testing.T) {
	tests := []struct {
		name          string
		registryName  string
		url           string
		config        map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid git registry",
			registryName: "my-git",
			url:          "https://github.com/user/repo",
			config:       map[string]string{"type": "git"},
			expectError:  false,
		},
		{
			name:         "valid s3 registry",
			registryName: "my-s3",
			url:          "my-bucket",
			config:       map[string]string{"type": "s3", "region": "us-east-1"},
			expectError:  false,
		},
		{
			name:          "missing config section",
			registryName:  "missing",
			url:           "test",
			config:        nil,
			expectError:   true,
			errorContains: "missing configuration section",
		},
		{
			name:          "missing type field",
			registryName:  "no-type",
			url:           "test",
			config:        map[string]string{},
			expectError:   true,
			errorContains: "missing required field 'type'",
		},
		{
			name:          "invalid registry type",
			registryName:  "invalid",
			url:           "test",
			config:        map[string]string{"type": "ftp"},
			expectError:   true,
			errorContains: "unknown registry type 'ftp'",
		},
		{
			name:          "s3 missing region",
			registryName:  "s3-no-region",
			url:           "my-bucket",
			config:        map[string]string{"type": "s3"},
			expectError:   true,
			errorContains: "missing required field 'region'",
		},
		{
			name:          "git missing url",
			registryName:  "git-no-url",
			url:           "",
			config:        map[string]string{"type": "git"},
			expectError:   true,
			errorContains: "missing registry URL",
		},
		{
			name:          "git non-https url",
			registryName:  "git-http",
			url:           "http://github.com/user/repo",
			config:        map[string]string{"type": "git"},
			expectError:   true,
			errorContains: "must use HTTPS protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegistry(tt.registryName, tt.url, tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateEngines(t *testing.T) {
	tests := []struct {
		name          string
		engines       map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:        "empty engines",
			engines:     map[string]string{},
			expectError: false,
		},
		{
			name:        "valid arm version",
			engines:     map[string]string{"arm": "^1.2.3"},
			expectError: false,
		},
		{
			name:        "valid arm version with tilde",
			engines:     map[string]string{"arm": "~1.2.3"},
			expectError: false,
		},
		{
			name:        "valid arm version exact",
			engines:     map[string]string{"arm": "1.2.3"},
			expectError: false,
		},
		{
			name:          "empty arm version",
			engines:       map[string]string{"arm": ""},
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "invalid arm version format",
			engines:       map[string]string{"arm": "invalid"},
			expectError:   true,
			errorContains: "invalid ARM engine version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEngines(tt.engines)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateChannels(t *testing.T) {
	tests := []struct {
		name          string
		channels      map[string]ChannelConfig
		expectError   bool
		errorContains string
	}{
		{
			name:        "empty channels",
			channels:    map[string]ChannelConfig{},
			expectError: false,
		},
		{
			name: "valid channels",
			channels: map[string]ChannelConfig{
				"cursor": {Directories: []string{"/path/to/cursor"}},
				"q":      {Directories: []string{"/path/to/q", "/another/path"}},
			},
			expectError: false,
		},
		{
			name: "channel with no directories",
			channels: map[string]ChannelConfig{
				"empty": {Directories: []string{}},
			},
			expectError:   true,
			errorContains: "must have at least one directory",
		},
		{
			name: "channel with empty directory",
			channels: map[string]ChannelConfig{
				"bad": {Directories: []string{"/valid", ""}},
			},
			expectError:   true,
			errorContains: "directory 1 cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannels(tt.channels)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	// Test valid configuration
	validCfg := &Config{
		Registries: map[string]string{
			"my-git": "https://github.com/user/repo",
			"my-s3":  "my-bucket",
		},
		RegistryConfigs: map[string]map[string]string{
			"my-git": {"type": "git"},
			"my-s3":  {"type": "s3", "region": "us-east-1"},
		},
		Engines: map[string]string{
			"arm": "^1.2.3",
		},
		Channels: map[string]ChannelConfig{
			"cursor": {Directories: []string{"/path/to/cursor"}},
		},
	}

	err := validateConfig(validCfg)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got: %v", err)
	}

	// Test invalid configuration
	invalidCfg := &Config{
		Registries: map[string]string{
			"bad-registry": "https://github.com/user/repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"bad-registry": {"type": "invalid-type"},
		},
	}

	err = validateConfig(invalidCfg)
	if err == nil {
		t.Error("Expected invalid config to fail validation")
	} else if !strings.Contains(err.Error(), "unknown registry type") {
		t.Errorf("Expected registry type error, got: %v", err)
	}
}

func TestGenerateStubFiles(t *testing.T) {
	// Test local stub generation
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(originalWd) }()

	// Generate local stubs
	err := GenerateStubFiles(false)
	if err != nil {
		t.Fatalf("Failed to generate local stub files: %v", err)
	}

	// Check that .armrc was created
	if _, err := os.Stat(".armrc"); os.IsNotExist(err) {
		t.Error("Expected .armrc stub file to be created")
	}

	// Check that arm.json was created
	if _, err := os.Stat("arm.json"); os.IsNotExist(err) {
		t.Error("Expected arm.json stub file to be created")
	}

	// Verify .armrc content
	armrcContent, err := os.ReadFile(".armrc")
	if err != nil {
		t.Fatalf("Failed to read .armrc: %v", err)
	}
	armrcStr := string(armrcContent)
	if !strings.Contains(armrcStr, "[registries]") {
		t.Error("Expected .armrc to contain [registries] section")
	}
	if !strings.Contains(armrcStr, "type = git") {
		t.Error("Expected .armrc to contain git type example")
	}
	if !strings.Contains(armrcStr, "$GITHUB_TOKEN") {
		t.Error("Expected .armrc to contain environment variable example")
	}

	// Verify arm.json content
	jsonContent, err := os.ReadFile("arm.json")
	if err != nil {
		t.Fatalf("Failed to read arm.json: %v", err)
	}
	jsonStr := string(jsonContent)
	if !strings.Contains(jsonStr, `"engines"`) {
		t.Error("Expected arm.json to contain engines section")
	}
	if !strings.Contains(jsonStr, `"arm": "^`) {
		t.Error("Expected arm.json to contain ARM version with ^ prefix")
	}
	if !strings.Contains(jsonStr, `"channels"`) {
		t.Error("Expected arm.json to contain channels section")
	}
	if !strings.Contains(jsonStr, `"rulesets"`) {
		t.Error("Expected arm.json to contain rulesets section")
	}

	// Test that files are not overwritten
	originalContent := string(armrcContent)
	err = GenerateStubFiles(false)
	if err != nil {
		t.Fatalf("Failed to run GenerateStubFiles again: %v", err)
	}
	newContent, _ := os.ReadFile(".armrc")
	if string(newContent) != originalContent {
		t.Error("Expected existing .armrc file to not be overwritten")
	}
}

func TestGenerateGlobalStubFiles(t *testing.T) {
	// Test global stub generation
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Generate global stubs
	err := GenerateStubFiles(true)
	if err != nil {
		t.Fatalf("Failed to generate global stub files: %v", err)
	}

	// Check that .arm directory was created
	armDir := filepath.Join(tmpDir, ".arm")
	if _, err := os.Stat(armDir); os.IsNotExist(err) {
		t.Error("Expected .arm directory to be created")
	}

	// Check that global .armrc was created
	globalARMRC := filepath.Join(armDir, ".armrc")
	if _, err := os.Stat(globalARMRC); os.IsNotExist(err) {
		t.Error("Expected global .armrc stub file to be created")
	}

	// Check that global arm.json was created
	globalJSON := filepath.Join(armDir, "arm.json")
	if _, err := os.Stat(globalJSON); os.IsNotExist(err) {
		t.Error("Expected global arm.json stub file to be created")
	}
}

func TestGenerateARMRCStub(t *testing.T) {
	tmpDir := t.TempDir()
	stubPath := filepath.Join(tmpDir, "test.armrc")

	err := generateARMRCStub(stubPath)
	if err != nil {
		t.Fatalf("Failed to generate .armrc stub: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(stubPath)
	if err != nil {
		t.Fatalf("Failed to stat stub file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("Expected file permissions 0o600, got %o", info.Mode().Perm())
	}

	// Check content structure
	content, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatalf("Failed to read stub file: %v", err)
	}

	contentStr := string(content)
	requiredSections := []string{
		"[registries]",
		"[registries.my-git-registry]",
		"[registries.my-s3-registry]",
		"[registries.my-gitlab-registry]",
		"[git]",
		"[https]",
		"[s3]",
		"[gitlab]",
		"[local]",
		"[network]",
		"[cache]",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("Expected stub to contain section %s", section)
		}
	}

	// Check for environment variable examples
	envVarExamples := []string{
		"$GITHUB_TOKEN",
		"$GITLAB_TOKEN",
	}

	for _, envVar := range envVarExamples {
		if !strings.Contains(contentStr, envVar) {
			t.Errorf("Expected stub to contain environment variable example %s", envVar)
		}
	}
}

func TestGenerateARMJSONStub(t *testing.T) {
	tmpDir := t.TempDir()
	stubPath := filepath.Join(tmpDir, "test.json")

	err := generateARMJSONStub(stubPath)
	if err != nil {
		t.Fatalf("Failed to generate arm.json stub: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(stubPath)
	if err != nil {
		t.Fatalf("Failed to stat stub file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("Expected file permissions 0o600, got %o", info.Mode().Perm())
	}

	// Check that it's valid JSON
	content, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatalf("Failed to read stub file: %v", err)
	}

	var armConfig ARMConfig
	if err := json.Unmarshal(content, &armConfig); err != nil {
		t.Errorf("Generated JSON is not valid: %v", err)
	}

	// Check required sections
	if armConfig.Engines == nil {
		t.Error("Expected engines section to be present")
	}
	if armConfig.Channels == nil {
		t.Error("Expected channels section to be present")
	}
	if armConfig.Rulesets == nil {
		t.Error("Expected rulesets section to be present")
	}

	// Check ARM version
	if armVersion, exists := armConfig.Engines["arm"]; !exists {
		t.Error("Expected ARM version to be present")
	} else if !strings.HasPrefix(armVersion, "^") {
		t.Errorf("Expected ARM version to start with ^, got %s", armVersion)
	}
}
