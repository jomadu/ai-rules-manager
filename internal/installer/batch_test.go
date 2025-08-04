package installer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchInstaller_InstallPackages(t *testing.T) {
	// Test empty packages
	batchInstaller := NewBatchInstaller(nil)
	result, err := batchInstaller.InstallPackages(map[string]string{})

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Successful)
	assert.Empty(t, result.Failed)
}

func TestBatchInstaller_NewBatchInstaller(t *testing.T) {
	// Test constructor
	batchInstaller := NewBatchInstaller(nil)
	assert.NotNil(t, batchInstaller)
}

func TestBatchInstaller_InstallFromManifest_InvalidFile(t *testing.T) {
	batchInstaller := NewBatchInstaller(nil)

	// Test with non-existent file
	result, err := batchInstaller.InstallFromManifest("nonexistent.json")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to load manifest")
}

func TestBatchInstaller_createBatchError(t *testing.T) {
	batchInstaller := NewBatchInstaller(nil)

	failures := map[string]error{
		"pkg1": errors.New("download failed"),
		"pkg2": errors.New("network timeout"),
	}

	err := batchInstaller.createBatchError(failures)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to install 2 package(s)")
	assert.Contains(t, err.Error(), "Check individual package errors")
}

func TestBatchResult_PrintResults(t *testing.T) {
	batchInstaller := NewBatchInstaller(nil)

	result := &BatchResult{
		Successful: []string{"pkg1", "pkg2"},
		Failed: map[string]error{
			"pkg3": errors.New("download failed"),
		},
	}

	// This test mainly ensures PrintResults doesn't panic
	// In a real scenario, you might want to capture stdout to verify output
	batchInstaller.PrintResults(result)
}
