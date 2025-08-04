package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryIntegration(t *testing.T) {
	env := NewTestEnv(t)

	t.Run("filesystem_registry", func(t *testing.T) {
		reg, err := env.RegistryManager.GetRegistry("filesystem")
		require.NoError(t, err)

		// Test health check
		err = reg.HealthCheck()
		assert.NoError(t, err)

		// Test list versions
		versions, err := reg.ListVersions("typescript-rules")
		require.NoError(t, err)
		assert.Contains(t, versions, "1.0.0")
	})

	t.Run("registry_concurrency", func(t *testing.T) {
		// Test concurrency settings
		defaultConcurrency := env.RegistryManager.GetConcurrency("default")
		assert.Equal(t, 3, defaultConcurrency) // Hardcoded default for generic type

		filesystemConcurrency := env.RegistryManager.GetConcurrency("filesystem")
		assert.Equal(t, 10, filesystemConcurrency) // Filesystem hardcoded default
	})

	t.Run("nonexistent_registry", func(t *testing.T) {
		_, err := env.RegistryManager.GetRegistry("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in configuration")
	})
}

func TestRegistryErrorHandling(t *testing.T) {
	env := NewTestEnv(t)

	t.Run("network_failure", func(t *testing.T) {
		// Close the HTTP server to simulate network failure
		env.HTTPServer.Close()

		reg, err := env.RegistryManager.GetRegistry("default")
		require.NoError(t, err)

		// Health check should fail
		err = reg.HealthCheck()
		assert.Error(t, err)

		// Download should fail
		_, err = reg.Download("typescript-rules", "1.0.0")
		assert.Error(t, err)
	})
}
