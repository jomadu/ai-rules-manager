package installer

import (
	"io"
	"strings"
	"testing"

	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
)

// MockRegistry for testing
type MockRegistry struct{}

func (m *MockRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	return &types.Ruleset{Name: name, Version: version}, nil
}

func (m *MockRegistry) ListVersions(name string) ([]string, error) {
	return []string{"1.0.0"}, nil
}

func (m *MockRegistry) Download(name, version string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("test data")), nil
}

func (m *MockRegistry) GetMetadata(name string) (*registry.Metadata, error) {
	return &registry.Metadata{Name: name}, nil
}

func TestInstaller(t *testing.T) {
	mockRegistry := &MockRegistry{}
	installer := New(mockRegistry)

	if installer.registry == nil {
		t.Error("Expected registry to be set")
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
