package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile_StructuredErrors(t *testing.T) {
	tests := []struct {
		name            string
		configContent   string
		expectedCode    errors.ErrorCode
		expectedInError []string
	}{
		{
			name: "invalid_syntax",
			configContent: `[sources
missing_bracket = invalid`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Failed to parse configuration file",
				"Check file syntax and format",
			},
		},
		{
			name: "negative_concurrency",
			configContent: `[sources]
test = http://example.com

[sources.test]
concurrency = -1`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Concurrency for source 'test' must be positive",
				"got -1",
			},
		},
		{
			name: "invalid_concurrency_value",
			configContent: `[sources]
test = http://example.com

[sources.test]
concurrency = not_a_number`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Invalid concurrency value for source 'test'",
				"not_a_number",
			},
		},
		{
			name: "negative_default_concurrency",
			configContent: `[sources]
test = http://example.com

[performance]
defaultConcurrency = -5`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Default concurrency must be positive",
				"got -5",
			},
		},
		{
			name: "invalid_default_concurrency",
			configContent: `[performance]
defaultConcurrency = invalid`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Invalid default concurrency value",
				"invalid",
			},
		},
		{
			name: "negative_performance_concurrency",
			configContent: `[performance.http]
concurrency = -2`,
			expectedCode: errors.ErrConfigInvalid,
			expectedInError: []string{
				"Performance concurrency for type 'http' must be positive",
				"got -2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".armrc")

			err := os.WriteFile(configPath, []byte(tt.configContent), 0o644)
			require.NoError(t, err)

			// Parse the config
			_, err = ParseFile(configPath)

			// Verify structured error
			require.Error(t, err)

			var armErr *errors.ARMError
			assert.ErrorAs(t, err, &armErr)
			assert.Equal(t, tt.expectedCode, armErr.Code)

			errorMsg := err.Error()
			for _, expected := range tt.expectedInError {
				assert.Contains(t, errorMsg, expected)
			}

			// Verify context is set
			assert.Equal(t, configPath, armErr.Context["file"])

			// Verify suggestions are present
			assert.NotEmpty(t, armErr.Suggestions)
		})
	}
}

func TestParseFile_ValidConfigs(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
	}{
		{
			name: "empty_concurrency_values",
			configContent: `[sources]
test = http://example.com

[sources.test]
concurrency =

[performance]
defaultConcurrency = `,
		},
		{
			name: "valid_concurrency_values",
			configContent: `[sources]
test = http://example.com

[sources.test]
concurrency = 5

[performance]
defaultConcurrency = 3

[performance.http]
concurrency = 4`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".armrc")

			err := os.WriteFile(configPath, []byte(tt.configContent), 0o644)
			require.NoError(t, err)

			// Parse the config - should succeed
			config, err := ParseFile(configPath)

			assert.NoError(t, err)
			assert.NotNil(t, config)
		})
	}
}

func TestParseFile_NonexistentFile(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/.armrc")

	require.Error(t, err)

	var armErr *errors.ARMError
	assert.ErrorAs(t, err, &armErr)
	assert.Equal(t, errors.ErrConfigInvalid, armErr.Code)
	assert.Contains(t, err.Error(), "Failed to parse configuration file")
	assert.Contains(t, err.Error(), "Ensure file exists and is readable")
}
