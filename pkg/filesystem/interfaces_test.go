package filesystem

import (
	"testing"
)

func TestAtomicFileSystemManager_Operations(t *testing.T) {
	manager := NewAtomicFileSystemManager("/tmp/test")

	files := []File{
		{Path: "test.md", Content: []byte("content")},
		{Path: "subdir/test2.md", Content: []byte("content2")},
	}

	// Test Install behavior - should create ARM directory structure
	err := manager.Install(".cursor/rules", "test-registry", "test-ruleset", "1.0.0", files)
	if err != nil {
		t.Errorf("install should not error for valid inputs: %v", err)
	}
	// TODO: Once implemented, verify directory structure: .cursor/rules/arm/test-registry/test-ruleset/1.0.0/

	// Test List behavior - should return installed file paths
	paths, err := manager.List(".cursor/rules", "test-registry", "test-ruleset", "1.0.0")
	if err != nil {
		t.Errorf("list should not error for valid inputs: %v", err)
	}
	// TODO: Once implemented, verify paths match installed files
	if paths == nil {
		t.Errorf("list should return non-nil slice")
	}

	// Test Uninstall behavior - should remove files atomically
	err = manager.Uninstall(".cursor/rules", "test-registry", "test-ruleset", "1.0.0")
	if err != nil {
		t.Errorf("uninstall should not error for valid inputs: %v", err)
	}
	// TODO: Once implemented, verify empty ARM directories are cleaned up

	// Test error cases
	_ = manager.Install("", "test-registry", "test-ruleset", "1.0.0", files)
	// TODO: Once implemented, verify error for invalid sink directory

	_ = manager.Install(".cursor/rules", "", "test-ruleset", "1.0.0", files)
	// TODO: Once implemented, verify error for empty registry name

	_ = manager.Install(".cursor/rules", "test-registry", "", "1.0.0", files)
	// TODO: Once implemented, verify error for empty ruleset name

	_ = manager.Install(".cursor/rules", "test-registry", "test-ruleset", "", files)
	// TODO: Once implemented, verify error for empty version
}

func TestFile_Structure(t *testing.T) {
	file := File{
		Path:    "rules/test.md",
		Content: []byte("# Test Rule\nThis is a test rule."),
	}

	if file.Path != "rules/test.md" {
		t.Errorf("expected path rules/test.md, got %s", file.Path)
	}

	expectedContent := "# Test Rule\nThis is a test rule."
	if string(file.Content) != expectedContent {
		t.Errorf("expected content %s, got %s", expectedContent, string(file.Content))
	}
}
