package types

import (
	"path/filepath"
	"testing"
)

func TestRulesManifestValidate(t *testing.T) {
	tests := []struct {
		name     string
		manifest RulesManifest
		wantErr  bool
	}{
		{
			"valid manifest",
			RulesManifest{
				Targets:      []string{".cursorrules", ".amazonq/rules"},
				Dependencies: map[string]string{"test-rules": "^1.0.0"},
			},
			false,
		},
		{
			"no targets",
			RulesManifest{
				Targets:      []string{},
				Dependencies: map[string]string{"test-rules": "^1.0.0"},
			},
			true,
		},
		{
			"nil dependencies",
			RulesManifest{
				Targets:      []string{".cursorrules"},
				Dependencies: nil,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RulesManifest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRulesLockValidate(t *testing.T) {
	tests := []struct {
		name    string
		lock    RulesLock
		wantErr bool
	}{
		{
			"valid lock",
			RulesLock{
				Version: "1",
				Dependencies: map[string]LockedDependency{
					"test-rules": {
						Version:  "1.0.0",
						Source:   "https://registry.example.com",
						Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
					},
				},
			},
			false,
		},
		{
			"missing version",
			RulesLock{
				Version: "",
				Dependencies: map[string]LockedDependency{
					"test-rules": {
						Version:  "1.0.0",
						Source:   "https://registry.example.com",
						Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
					},
				},
			},
			true,
		},
		{
			"invalid dependency checksum",
			RulesLock{
				Version: "1",
				Dependencies: map[string]LockedDependency{
					"test-rules": {
						Version:  "1.0.0",
						Source:   "https://registry.example.com",
						Checksum: "invalid",
					},
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lock.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RulesLock.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadSaveManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "rules.json")

	original := &RulesManifest{
		Targets:      []string{".cursorrules", ".amazonq/rules"},
		Dependencies: map[string]string{"test-rules": "^1.0.0"},
	}

	// Save manifest
	if err := original.SaveManifest(manifestPath); err != nil {
		t.Fatalf("SaveManifest() error = %v", err)
	}

	// Load manifest
	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}

	// Compare
	if len(loaded.Targets) != len(original.Targets) {
		t.Errorf("Targets length mismatch: got %d, want %d", len(loaded.Targets), len(original.Targets))
	}
	if len(loaded.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length mismatch: got %d, want %d", len(loaded.Dependencies), len(original.Dependencies))
	}
}

func TestLoadSaveLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "rules.lock")

	original := &RulesLock{
		Version: "1",
		Dependencies: map[string]LockedDependency{
			"test-rules": {
				Version:  "1.0.0",
				Source:   "https://registry.example.com",
				Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			},
		},
	}

	// Save lock file
	if err := original.SaveLockFile(lockPath); err != nil {
		t.Fatalf("SaveLockFile() error = %v", err)
	}

	// Load lock file
	loaded, err := LoadLockFile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockFile() error = %v", err)
	}

	// Compare
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: got %s, want %s", loaded.Version, original.Version)
	}
	if len(loaded.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length mismatch: got %d, want %d", len(loaded.Dependencies), len(original.Dependencies))
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := LoadManifest("nonexistent.json")
	if err == nil {
		t.Error("LoadManifest() should return error for nonexistent file")
	}

	_, err = LoadLockFile("nonexistent.lock")
	if err == nil {
		t.Error("LoadLockFile() should return error for nonexistent file")
	}
}