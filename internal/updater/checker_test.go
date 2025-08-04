package updater

import (
	"testing"

	"github.com/jomadu/arm/internal/registry"
)

func TestCheckStatus_String(t *testing.T) {
	tests := []struct {
		status   CheckStatus
		expected string
	}{
		{CheckUpToDate, "Up to date"},
		{CheckOutdated, "Update available"},
		{CheckError, "Error"},
		{CheckNoCompatible, "No compatible update"},
		{CheckStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("CheckStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewChecker(t *testing.T) {
	manager := &registry.Manager{}
	checker := NewChecker(manager)

	if checker == nil {
		t.Fatal("NewChecker() returned nil")
	}

	if checker.manager != manager {
		t.Error("NewChecker() did not set manager correctly")
	}
}

func TestCheckRuleset_InvalidCurrentVersion(t *testing.T) {
	manager := &registry.Manager{}
	checker := NewChecker(manager)

	ruleset := InstalledRuleset{
		Name:    "test-ruleset",
		Version: "invalid-version",
		Source:  "test-source",
	}

	result := checker.CheckRuleset(ruleset, "^1.0.0")

	if result.Status != CheckError {
		t.Errorf("Expected CheckError, got %v", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error for invalid version")
	}

	if result.Name != "test-ruleset" {
		t.Errorf("Expected name 'test-ruleset', got %v", result.Name)
	}

	if result.Current != "invalid-version" {
		t.Errorf("Expected current 'invalid-version', got %v", result.Current)
	}
}

func TestCheckRuleset_InvalidConstraint(t *testing.T) {
	manager := &registry.Manager{}
	checker := NewChecker(manager)

	ruleset := InstalledRuleset{
		Name:    "test-ruleset",
		Version: "1.0.0",
		Source:  "test-source",
	}

	result := checker.CheckRuleset(ruleset, "invalid-constraint")

	if result.Status != CheckError {
		t.Errorf("Expected CheckError, got %v", result.Status)
	}

	if result.Error == nil {
		t.Error("Expected error for invalid constraint")
	}

	if result.Constraint != "invalid-constraint" {
		t.Errorf("Expected constraint 'invalid-constraint', got %v", result.Constraint)
	}
}
