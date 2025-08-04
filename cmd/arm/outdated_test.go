package main

import (
	"testing"

	"github.com/jomadu/arm/internal/updater"
)

func TestFilterOutdated(t *testing.T) {
	results := []updater.CheckResult{
		{Name: "outdated1", Status: updater.CheckOutdated},
		{Name: "uptodate1", Status: updater.CheckUpToDate},
		{Name: "outdated2", Status: updater.CheckOutdated},
		{Name: "error1", Status: updater.CheckError},
	}

	filtered := filterOutdated(results)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 outdated results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.Status != updater.CheckOutdated {
			t.Errorf("Expected only outdated results, got %v", result.Status)
		}
	}

	expectedNames := map[string]bool{"outdated1": true, "outdated2": true}
	for _, result := range filtered {
		if !expectedNames[result.Name] {
			t.Errorf("Unexpected result name: %s", result.Name)
		}
	}
}

func TestFilterOutdated_Empty(t *testing.T) {
	results := []updater.CheckResult{
		{Name: "uptodate1", Status: updater.CheckUpToDate},
		{Name: "error1", Status: updater.CheckError},
	}

	filtered := filterOutdated(results)

	if len(filtered) != 0 {
		t.Errorf("Expected 0 outdated results, got %d", len(filtered))
	}
}

func TestFilterOutdated_AllOutdated(t *testing.T) {
	results := []updater.CheckResult{
		{Name: "outdated1", Status: updater.CheckOutdated},
		{Name: "outdated2", Status: updater.CheckOutdated},
	}

	filtered := filterOutdated(results)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 outdated results, got %d", len(filtered))
	}
}
