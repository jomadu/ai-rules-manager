package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCacheManager(t *testing.T) {
	cm, err := NewCacheManager()
	if err != nil {
		t.Fatalf("NewCacheManager() error = %v", err)
	}
	
	if cm.CacheDir == "" {
		t.Error("CacheManager should have non-empty CacheDir")
	}
	
	// Check that cache directory exists
	if _, err := os.Stat(cm.CacheDir); os.IsNotExist(err) {
		t.Error("Cache directory should be created")
	}
}

func TestGetRulesetCachePath(t *testing.T) {
	cm := &CacheManager{CacheDir: "/test/cache"}
	
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
			"/test/cache/typescript-rules/1.0.0",
		},
		{
			"scoped package",
			"company@security-rules",
			"2.1.0",
			"/test/cache/@company/security-rules/2.1.0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.GetRulesetCachePath(tt.rulesetName, tt.version)
			if got != tt.want {
				t.Errorf("GetRulesetCachePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPackagePath(t *testing.T) {
	cm := &CacheManager{CacheDir: "/test/cache"}
	
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
			"/test/cache/typescript-rules/1.0.0/package.tar.gz",
		},
		{
			"scoped package",
			"company@security-rules",
			"2.1.0",
			"/test/cache/@company/security-rules/2.1.0/package.tar.gz",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.GetPackagePath(tt.rulesetName, tt.version)
			if got != tt.want {
				t.Errorf("GetPackagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCached(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &CacheManager{CacheDir: tmpDir}
	
	rulesetName := "test-rules"
	version := "1.0.0"
	
	// Should not be cached initially
	if cm.IsCached(rulesetName, version) {
		t.Error("Ruleset should not be cached initially")
	}
	
	// Create the package file
	packagePath := cm.GetPackagePath(rulesetName, version)
	if err := os.MkdirAll(filepath.Dir(packagePath), 0755); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}
	if err := os.WriteFile(packagePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create package file: %v", err)
	}
	
	// Should be cached now
	if !cm.IsCached(rulesetName, version) {
		t.Error("Ruleset should be cached after creating package file")
	}
}

func TestEnsureCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &CacheManager{CacheDir: tmpDir}
	
	rulesetName := "test-rules"
	version := "1.0.0"
	
	if err := cm.EnsureCacheDir(rulesetName, version); err != nil {
		t.Fatalf("EnsureCacheDir() error = %v", err)
	}
	
	// Check that directory was created
	cachePath := cm.GetRulesetCachePath(rulesetName, version)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Cache directory should be created")
	}
}

func TestGetTargetPath(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		rulesetName string
		version     string
		want        string
	}{
		{
			"unscoped package",
			".cursorrules",
			"typescript-rules",
			"1.0.0",
			".cursorrules/arm/typescript-rules/1.0.0",
		},
		{
			"scoped package",
			".amazonq/rules",
			"company@security-rules",
			"2.1.0",
			".amazonq/rules/arm/@company/security-rules/2.1.0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTargetPath(tt.target, tt.rulesetName, tt.version)
			if got != tt.want {
				t.Errorf("GetTargetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}