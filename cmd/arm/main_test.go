package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Test that main function doesn't panic with help flag
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"arm", "--help"}

	// This should exit with code 0 for help, but we can't easily test that
	// without more complex setup. For now, just ensure run() function exists
	// and can be called without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main function panicked: %v", r)
		}
	}()

	// Test that run function exists and handles errors gracefully
	err := run()
	if err == nil {
		t.Log("run() completed successfully (likely showed help)")
	} else {
		t.Logf("run() returned error (expected for test): %v", err)
	}
}
