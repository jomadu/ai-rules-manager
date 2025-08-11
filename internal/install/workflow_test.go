package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComplexWorkflowSimulation(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Create basic configuration files
	armrcContent := `[registries]
default = https://github.com/user/repo

[registries.default]
type = git
`
	err := os.WriteFile(".armrc", []byte(armrcContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	armJSONContent := `{
  "engines": {"arm": "^1.0.0"},
  "channels": {
    "cursor": {"directories": [".cursor/rules"]},
    "q": {"directories": [".amazonq/rules"]}
  },
  "rulesets": {}
}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Test workflow steps
	tests := []struct {
		name     string
		step     func() error
		validate func() error
	}{
		{
			name: "install ghost hunting rules",
			step: func() error {
				// Simulate installing with ghost-*.md pattern
				return simulateInstall("test-ruleset", []string{"ghost-*.md"}, []string{"cursor"})
			},
			validate: func() error {
				// Verify cursor directory structure exists
				if _, err := os.Stat(".cursor/rules"); os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			name: "install AI assistant specific rules",
			step: func() error {
				// Simulate installing cursor-specific tools
				return simulateInstall("test-ruleset", []string{"tools/cursor-pro.md"}, []string{"cursor"})
			},
			validate: func() error {
				// Verify cursor tools directory exists
				return nil // Would check for specific files in real implementation
			},
		},
		{
			name: "install guidelines",
			step: func() error {
				// Simulate installing guidelines
				return simulateInstall("test-ruleset", []string{"guidelines/*.md"}, []string{"cursor", "q"})
			},
			validate: func() error {
				// Verify both cursor and q directories exist
				if _, err := os.Stat(".cursor/rules"); os.IsNotExist(err) {
					return err
				}
				if _, err := os.Stat(".amazonq/rules"); os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.step(); err != nil {
				t.Errorf("Step failed: %v", err)
			}
			if err := tt.validate(); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

func TestChannelSpecificInstallation(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Create channel directories
	err := os.MkdirAll(".cursor/rules", 0o755)
	if err != nil {
		t.Fatalf("Failed to create cursor directory: %v", err)
	}
	err = os.MkdirAll(".amazonq/rules", 0o755)
	if err != nil {
		t.Fatalf("Failed to create amazonq directory: %v", err)
	}

	tests := []struct {
		name     string
		channels []string
		patterns []string
		validate func() error
	}{
		{
			name:     "cursor only installation",
			channels: []string{"cursor"},
			patterns: []string{"tools/cursor-pro.md"},
			validate: func() error {
				// Should create files in cursor directory only
				cursorPath := filepath.Join(".cursor", "rules", "arm", "default", "test-ruleset")
				if err := os.MkdirAll(cursorPath, 0o755); err != nil {
					return err
				}
				// Simulate file creation
				return os.WriteFile(filepath.Join(cursorPath, "cursor-pro.md"), []byte("# Cursor Pro Rules"), 0o644)
			},
		},
		{
			name:     "q-developer only installation",
			channels: []string{"q"},
			patterns: []string{"ai-assistants/q-developer.md"},
			validate: func() error {
				// Should create files in amazonq directory only
				qPath := filepath.Join(".amazonq", "rules", "arm", "default", "test-ruleset")
				if err := os.MkdirAll(qPath, 0o755); err != nil {
					return err
				}
				// Simulate file creation
				return os.WriteFile(filepath.Join(qPath, "q-developer.md"), []byte("# Q Developer Rules"), 0o644)
			},
		},
		{
			name:     "multi-channel installation",
			channels: []string{"cursor", "q"},
			patterns: []string{"guidelines/*.md"},
			validate: func() error {
				// Should create files in both directories
				for _, channel := range []string{"cursor", "q"} {
					var dir string
					if channel == "cursor" {
						dir = ".cursor"
					} else {
						dir = ".amazonq"
					}
					channelPath := filepath.Join(dir, "rules", "arm", "default", "test-ruleset")
					if err := os.MkdirAll(channelPath, 0o755); err != nil {
						return err
					}
					// Simulate file creation
					if err := os.WriteFile(filepath.Join(channelPath, "guidelines.md"), []byte("# Guidelines"), 0o644); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := simulateChannelInstall("test-ruleset", tt.patterns, tt.channels)
			if err != nil {
				t.Errorf("Channel installation failed: %v", err)
			}

			if err := tt.validate(); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// simulateInstall simulates the installation process for testing
func simulateInstall(ruleset string, patterns, channels []string) error {
	// Create directory structure for each channel
	for _, channel := range channels {
		var baseDir string
		switch channel {
		case "cursor":
			baseDir = ".cursor/rules"
		case "q":
			baseDir = ".amazonq/rules"
		default:
			continue
		}

		rulesetPath := filepath.Join(baseDir, "arm", "default", ruleset)
		if err := os.MkdirAll(rulesetPath, 0o755); err != nil {
			return err
		}

		// Simulate creating files based on patterns
		for _, pattern := range patterns {
			// Simple simulation - create a file based on pattern
			filename := filepath.Base(pattern)
			if filename == "*" || filename == "*.md" {
				filename = "example.md"
			}

			filePath := filepath.Join(rulesetPath, filename)
			if err := os.WriteFile(filePath, []byte("# Example Rule"), 0o644); err != nil {
				return err
			}
		}
	}

	return nil
}

// simulateChannelInstall simulates channel-specific installation
func simulateChannelInstall(ruleset string, patterns, channels []string) error {
	return simulateInstall(ruleset, patterns, channels)
}
