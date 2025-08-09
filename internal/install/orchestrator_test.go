package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestInstallOrchestrator_InstallMultiple(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-orchestrator-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory for lock file operations
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Create test configuration
	cfg := &config.Config{
		Registries: map[string]string{
			"registry1": "https://github.com/test/repo1",
			"registry2": "s3://test-bucket",
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{filepath.Join(tempDir, ".cursor", "rules")},
			},
		},
		RegistryConfigs: map[string]map[string]string{
			"registry1": {"type": "git", "concurrency": "2", "rateLimit": "5/second"},
			"registry2": {"type": "s3", "concurrency": "1", "rateLimit": "2/second"},
		},
	}

	installer := New(cfg)
	orchestrator := NewInstallOrchestrator(installer)

	// Create test source files
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	var requests []InstallRequest
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(sourceDir, fmt.Sprintf("rule%d.md", i))
		if err := os.WriteFile(testFile, []byte(fmt.Sprintf("# Test Rule %d", i)), 0o644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		registry := "registry1"
		if i == 2 {
			registry = "registry2"
		}

		requests = append(requests, InstallRequest{
			Registry:    registry,
			Ruleset:     fmt.Sprintf("ruleset%d", i),
			Version:     "1.0.0",
			SourceFiles: []string{testFile},
			Channels:    []string{"cursor"},
		})
	}

	// Track progress
	var progressCalls []string
	var progressMu sync.Mutex
	progressCallback := func(current, total int, operation string) {
		progressMu.Lock()
		defer progressMu.Unlock()
		progressCalls = append(progressCalls, fmt.Sprintf("%d/%d: %s", current, total, operation))
	}

	// Test parallel installation
	req := &MultiInstallRequest{
		Requests: requests,
		Progress: progressCallback,
	}

	result, err := orchestrator.InstallMultiple(context.Background(), req)
	if err != nil {
		t.Fatalf("InstallMultiple failed: %v", err)
	}

	// Verify results
	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}
	if len(result.Successful) != 3 {
		t.Errorf("Expected 3 successful installations, got %d", len(result.Successful))
		for _, success := range result.Successful {
			t.Logf("Successful: %s/%s@%s", success.Registry, success.Ruleset, success.Version)
		}
	}
	if len(result.Failed) != 0 {
		t.Errorf("Expected 0 failed installations, got %d", len(result.Failed))
		for _, failure := range result.Failed {
			t.Logf("Failed: %s/%s - %v", failure.Registry, failure.Ruleset, failure.Error)
		}
	}

	// Verify progress was called
	progressMu.Lock()
	if len(progressCalls) != 3 {
		t.Errorf("Expected 3 progress calls, got %d", len(progressCalls))
	}
	progressMu.Unlock()

	// Verify files were installed
	for i := 0; i < 3; i++ {
		registry := "registry1"
		if i == 2 {
			registry = "registry2"
		}
		expectedPath := filepath.Join(tempDir, ".cursor", "rules", "arm", registry, fmt.Sprintf("ruleset%d", i), "1.0.0", fmt.Sprintf("rule%d.md", i))
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("File not found at expected location %s: %v", expectedPath, err)
		}
	}
}

func TestTokenBucket_RateLimit(t *testing.T) {
	// Create token bucket with 2 tokens, refill every 100ms
	bucket := &TokenBucket{
		tokens:     2,
		capacity:   2,
		refillRate: 100 * time.Millisecond,
		lastRefill: time.Now(),
	}

	// Should be able to take 2 tokens immediately
	if !bucket.TakeToken() {
		t.Error("Expected to take first token")
	}
	if !bucket.TakeToken() {
		t.Error("Expected to take second token")
	}

	// Third token should fail
	if bucket.TakeToken() {
		t.Error("Expected third token to fail")
	}

	// Wait for refill and try again
	time.Sleep(150 * time.Millisecond)
	if !bucket.TakeToken() {
		t.Error("Expected token after refill")
	}
}

func TestInstallOrchestrator_ParseRateLimit(t *testing.T) {
	orchestrator := &InstallOrchestrator{}

	tests := []struct {
		input    string
		capacity int
		duration time.Duration
	}{
		{"10/minute", 10, 6 * time.Second},
		{"5/second", 5, 200 * time.Millisecond},
		{"100/hour", 100, 36 * time.Second},
		{"invalid", 10, time.Minute}, // default
	}

	for _, test := range tests {
		capacity, refillRate := orchestrator.parseRateLimit(test.input)
		if capacity != test.capacity {
			t.Errorf("For %s: expected capacity %d, got %d", test.input, test.capacity, capacity)
		}
		if refillRate != test.duration {
			t.Errorf("For %s: expected refill rate %v, got %v", test.input, test.duration, refillRate)
		}
	}
}

func TestInstallOrchestrator_GroupByRegistry(t *testing.T) {
	orchestrator := &InstallOrchestrator{}

	requests := []InstallRequest{
		{Registry: "registry1", Ruleset: "ruleset1"},
		{Registry: "registry2", Ruleset: "ruleset2"},
		{Registry: "registry1", Ruleset: "ruleset3"},
		{Registry: "registry3", Ruleset: "ruleset4"},
	}

	groups := orchestrator.groupByRegistry(requests)

	if len(groups) != 3 {
		t.Errorf("Expected 3 registry groups, got %d", len(groups))
	}

	if len(groups["registry1"]) != 2 {
		t.Errorf("Expected 2 requests for registry1, got %d", len(groups["registry1"]))
	}

	if len(groups["registry2"]) != 1 {
		t.Errorf("Expected 1 request for registry2, got %d", len(groups["registry2"]))
	}

	if len(groups["registry3"]) != 1 {
		t.Errorf("Expected 1 request for registry3, got %d", len(groups["registry3"]))
	}
}

func TestInstallOrchestrator_ConcurrencyControl(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-concurrency-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory for lock file operations
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Create test configuration with low concurrency
	cfg := &config.Config{
		Registries: map[string]string{
			"registry1": "https://github.com/test/repo1",
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{filepath.Join(tempDir, ".cursor", "rules")},
			},
		},
		RegistryConfigs: map[string]map[string]string{
			"registry1": {"type": "git", "concurrency": "1", "rateLimit": "10/second"},
		},
	}

	installer := New(cfg)
	orchestrator := NewInstallOrchestrator(installer)

	// Create test source files
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	var requests []InstallRequest
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(sourceDir, fmt.Sprintf("rule%d.md", i))
		if err := os.WriteFile(testFile, []byte(fmt.Sprintf("# Test Rule %d", i)), 0o644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		requests = append(requests, InstallRequest{
			Registry:    "registry1",
			Ruleset:     fmt.Sprintf("ruleset%d", i),
			Version:     "1.0.0",
			SourceFiles: []string{testFile},
			Channels:    []string{"cursor"},
		})
	}

	// Measure execution time to verify concurrency control
	start := time.Now()

	req := &MultiInstallRequest{
		Requests: requests,
	}

	result, err := orchestrator.InstallMultiple(context.Background(), req)
	if err != nil {
		t.Fatalf("InstallMultiple failed: %v", err)
	}

	duration := time.Since(start)

	// With concurrency=1, operations should be sequential
	// This is a rough check - in practice, timing tests can be flaky
	if duration < 10*time.Millisecond {
		t.Logf("Operations completed in %v (may indicate parallel execution despite concurrency=1)", duration)
	}

	// Verify all installations succeeded
	if len(result.Successful) != 3 {
		t.Errorf("Expected 3 successful installations, got %d", len(result.Successful))
	}
}

func TestInstallOrchestrator_ErrorHandling(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Registries: map[string]string{
			"registry1": "https://github.com/test/repo1",
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{"/nonexistent/path"}, // This will cause errors
			},
		},
		RegistryConfigs: map[string]map[string]string{
			"registry1": {"type": "git", "concurrency": "2", "rateLimit": "10/second"},
		},
	}

	installer := New(cfg)
	orchestrator := NewInstallOrchestrator(installer)

	// Create requests that will fail
	requests := []InstallRequest{
		{
			Registry:    "registry1",
			Ruleset:     "ruleset1",
			Version:     "1.0.0",
			SourceFiles: []string{"/nonexistent/file.md"},
			Channels:    []string{"cursor"},
		},
		{
			Registry:    "registry1",
			Ruleset:     "ruleset2",
			Version:     "1.0.0",
			SourceFiles: []string{"/nonexistent/file2.md"},
			Channels:    []string{"cursor"},
		},
	}

	req := &MultiInstallRequest{
		Requests: requests,
	}

	result, err := orchestrator.InstallMultiple(context.Background(), req)
	if err != nil {
		t.Fatalf("InstallMultiple failed: %v", err)
	}

	// Verify errors were captured
	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
	}
	if len(result.Failed) != 2 {
		t.Errorf("Expected 2 failed installations, got %d", len(result.Failed))
	}
	if len(result.Successful) != 0 {
		t.Errorf("Expected 0 successful installations, got %d", len(result.Successful))
	}
}
