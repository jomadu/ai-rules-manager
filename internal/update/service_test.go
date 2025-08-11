package update

import (
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestParseRulesetSpec(t *testing.T) {
	tests := []struct {
		spec             string
		expectedRegistry string
		expectedName     string
		expectedVersion  string
	}{
		{"my-rules", "", "my-rules", "latest"},
		{"registry/my-rules", "registry", "my-rules", "latest"},
		{"my-rules@1.0.0", "", "my-rules", "1.0.0"},
		{"registry/my-rules@1.0.0", "registry", "my-rules", "1.0.0"},
	}

	for _, tt := range tests {
		registry, name, version := parseRulesetSpec(tt.spec)
		if registry != tt.expectedRegistry {
			t.Errorf("parseRulesetSpec(%q) registry = %q, want %q", tt.spec, registry, tt.expectedRegistry)
		}
		if name != tt.expectedName {
			t.Errorf("parseRulesetSpec(%q) name = %q, want %q", tt.spec, name, tt.expectedName)
		}
		if version != tt.expectedVersion {
			t.Errorf("parseRulesetSpec(%q) version = %q, want %q", tt.spec, version, tt.expectedVersion)
		}
	}
}

func TestUpdateService_New(t *testing.T) {
	cfg := &config.Config{}
	service := New(cfg)

	if service == nil {
		t.Error("New() returned nil")
		return
	}

	if service.config != cfg {
		t.Error("New() did not set config correctly")
	}
}
