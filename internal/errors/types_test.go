package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestARMError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ARMError
		expected []string // Parts that should be in the error message
	}{
		{
			name: "basic_error",
			err:  New(ErrPackageNotFound, "Package not found"),
			expected: []string{
				"[PACKAGE_NOT_FOUND] Package not found",
			},
		},
		{
			name: "error_with_context",
			err: New(ErrPackageNotFound, "Package not found").
				WithContext("package", "test-pkg").
				WithContext("version", "1.0.0"),
			expected: []string{
				"[PACKAGE_NOT_FOUND] Package not found",
				"Context:",
				"package=test-pkg",
				"version=1.0.0",
			},
		},
		{
			name: "error_with_suggestions",
			err: New(ErrPackageNotFound, "Package not found").
				WithSuggestion("Check package name").
				WithSuggestion("Verify registry"),
			expected: []string{
				"[PACKAGE_NOT_FOUND] Package not found",
				"Suggestions:",
				"  - Check package name",
				"  - Verify registry",
			},
		},
		{
			name: "error_with_cause",
			err:  Wrap(errors.New("network timeout"), ErrNetworkTimeout, "Connection failed"),
			expected: []string{
				"[NETWORK_TIMEOUT] Connection failed",
				"Caused by: network timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestARMError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := Wrap(cause, ErrNetworkTimeout, "Wrapped error")

	assert.Equal(t, cause, err.Unwrap())
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("NetworkTimeout", func(t *testing.T) {
		err := NetworkTimeout("test-registry")
		assert.Equal(t, ErrNetworkTimeout, err.Code)
		assert.Contains(t, err.Error(), "test-registry")
		assert.Contains(t, err.Error(), "Check your internet connection")
	})

	t.Run("PackageNotFound", func(t *testing.T) {
		err := PackageNotFound("test-pkg", "1.0.0", "test-registry")
		assert.Equal(t, ErrPackageNotFound, err.Code)
		assert.Contains(t, err.Error(), "test-pkg@1.0.0")
		assert.Contains(t, err.Error(), "test-registry")
		assert.Contains(t, err.Error(), "Check package name spelling")
	})

	t.Run("ConfigInvalid", func(t *testing.T) {
		err := ConfigInvalid("/path/to/config", "syntax error")
		assert.Equal(t, ErrConfigInvalid, err.Code)
		assert.Contains(t, err.Error(), "/path/to/config")
		assert.Contains(t, err.Error(), "syntax error")
	})

	t.Run("SourceNotFound", func(t *testing.T) {
		err := SourceNotFound("test-source")
		assert.Equal(t, ErrSourceNotFound, err.Code)
		assert.Contains(t, err.Error(), "test-source")
		assert.Contains(t, err.Error(), "arm config set sources.test-source")
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		err := PermissionDenied("/path/to/file")
		assert.Equal(t, ErrPermissionDenied, err.Code)
		assert.Contains(t, err.Error(), "/path/to/file")
		assert.Contains(t, err.Error(), "Check file/directory permissions")
	})
}

func TestARMError_Chaining(t *testing.T) {
	err := New(ErrPackageNotFound, "Package not found").
		WithContext("package", "test").
		WithContext("version", "1.0.0").
		WithSuggestion("Check spelling").
		WithSuggestion("Verify registry")

	result := err.Error()

	// Should contain all parts
	assert.Contains(t, result, "[PACKAGE_NOT_FOUND]")
	assert.Contains(t, result, "Package not found")
	assert.Contains(t, result, "package=test")
	assert.Contains(t, result, "version=1.0.0")
	assert.Contains(t, result, "Check spelling")
	assert.Contains(t, result, "Verify registry")
}

func TestErrorCodes(t *testing.T) {
	// Test that error codes are properly defined
	codes := []ErrorCode{
		ErrNetworkTimeout,
		ErrNetworkUnreachable,
		ErrPackageNotFound,
		ErrRegistryNotFound,
		ErrConfigInvalid,
		ErrConfigMissing,
		ErrSourceNotFound,
		ErrManifestInvalid,
		ErrPermissionDenied,
		ErrDiskSpace,
		ErrFileNotFound,
		ErrVersionInvalid,
		ErrVersionNotFound,
		ErrVersionConflict,
	}

	for _, code := range codes {
		assert.NotEmpty(t, string(code), "Error code should not be empty")
		assert.True(t, strings.ToUpper(string(code)) == string(code), "Error code should be uppercase")
	}
}
