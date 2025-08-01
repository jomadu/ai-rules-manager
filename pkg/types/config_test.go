package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildDownloadURL(t *testing.T) {
	config := &RegistryConfig{
		Sources: map[string]RegistrySource{
			"default": {URL: "https://registry.armjs.org/"},
			"company": {URL: "https://internal.company.com/"},
		},
	}

	tests := []struct {
		name        string
		rulesetName string
		version     string
		want        string
	}{
		{
			"unscoped package",
			"typescript-rules",
			"1.0.0",
			"https://registry.armjs.org/typescript-rules/1.0.0/package.tgz",
		},
		{
			"scoped package with matching source",
			"company@security-rules",
			"2.1.0",
			"https://internal.company.com/@company/security-rules/2.1.0/package.tgz",
		},
		{
			"scoped package without matching source",
			"other@test-rules",
			"1.0.0",
			"https://registry.armjs.org/@other/test-rules/1.0.0/package.tgz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.BuildDownloadURL(tt.rulesetName, tt.version)
			if err != nil {
				t.Errorf("BuildDownloadURL() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("BuildDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveRulesetSource(t *testing.T) {
	config := &RegistryConfig{
		Sources: map[string]RegistrySource{
			"default": {URL: "https://registry.armjs.org/"},
			"company": {URL: "https://internal.company.com/"},
		},
	}

	tests := []struct {
		name        string
		rulesetName string
		want        string
	}{
		{"unscoped package", "typescript-rules", "default"},
		{"scoped package with matching source", "company@security-rules", "company"},
		{"scoped package without matching source", "other@test-rules", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.ResolveRulesetSource(tt.rulesetName)
			if got != tt.want {
				t.Errorf("ResolveRulesetSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRegistryURL(t *testing.T) {
	config := &RegistryConfig{
		Sources: map[string]RegistrySource{
			"default": {URL: "https://registry.armjs.org/"},
			"company": {URL: "https://internal.company.com/"},
		},
	}

	tests := []struct {
		name       string
		sourceName string
		want       string
		wantErr    bool
	}{
		{"existing source", "default", "https://registry.armjs.org", false},
		{"existing source with trailing slash", "company", "https://internal.company.com", false},
		{"nonexistent source", "nonexistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.GetRegistryURL(tt.sourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegistryURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRegistryURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAuthToken(t *testing.T) {
	config := &RegistryConfig{
		Sources: map[string]RegistrySource{
			"default": {URL: "https://registry.armjs.org/"},
			"company": {URL: "https://internal.company.com/", AuthToken: "secret-token"},
		},
	}

	tests := []struct {
		name       string
		sourceName string
		want       string
	}{
		{"source without token", "default", ""},
		{"source with token", "company", "secret-token"},
		{"nonexistent source", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetAuthToken(tt.sourceName)
			if got != tt.want {
				t.Errorf("GetAuthToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadRegistryConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".armrc")

	// Create test config file
	configContent := `[sources]
default = https://registry.armjs.org/
company = https://internal.company.com/

[sources.company]
authToken = test-token
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Change to temp directory to test project-level config loading
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config, err := LoadRegistryConfig()
	if err != nil {
		t.Fatalf("LoadRegistryConfig() error = %v", err)
	}

	// Check default source exists
	if _, exists := config.Sources["default"]; !exists {
		t.Error("Default source should exist")
	}

	// Check company source
	companySource, exists := config.Sources["company"]
	if !exists {
		t.Error("Company source should exist")
	}
	if companySource.AuthToken != "test-token" {
		t.Errorf("Company auth token = %v, want test-token", companySource.AuthToken)
	}
}
