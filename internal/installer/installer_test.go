package installer

import (
	"testing"

	"github.com/jomadu/arm/pkg/types"
)

func TestBuildDownloadURL(t *testing.T) {
	installer := New("https://registry.example.com")

	tests := []struct {
		name     string
		org      string
		pkg      string
		version  string
		expected string
	}{
		{
			name:     "package without org",
			org:      "",
			pkg:      "typescript-rules",
			version:  "1.0.0",
			expected: "https://registry.example.com/typescript-rules/1.0.0.tar.gz",
		},
		{
			name:     "package with org",
			org:      "company",
			pkg:      "typescript-rules",
			version:  "1.0.0",
			expected: "https://registry.example.com/company/typescript-rules/1.0.0.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installer.buildDownloadURL(tt.org, tt.pkg, tt.version)
			if result != tt.expected {
				t.Errorf("buildDownloadURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseRulesetName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedOrg string
		expectedPkg string
	}{
		{
			name:        "package without org",
			input:       "typescript-rules",
			expectedOrg: "",
			expectedPkg: "typescript-rules",
		},
		{
			name:        "package with org",
			input:       "company@typescript-rules",
			expectedOrg: "company",
			expectedPkg: "typescript-rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, pkg := types.ParseRulesetName(tt.input)
			if org != tt.expectedOrg {
				t.Errorf("ParseRulesetName() org = %v, want %v", org, tt.expectedOrg)
			}
			if pkg != tt.expectedPkg {
				t.Errorf("ParseRulesetName() pkg = %v, want %v", pkg, tt.expectedPkg)
			}
		})
	}
}
