package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedType     string
		expectedGraceful bool
	}{
		{
			name:             "network error",
			err:              &NetworkError{Message: "connection timeout"},
			expectedType:     "network",
			expectedGraceful: true,
		},
		{
			name:             "registry not found",
			err:              errors.New("registry 'nonexistent' not found"),
			expectedType:     "registry",
			expectedGraceful: true,
		},
		{
			name:             "invalid pattern",
			err:              errors.New("invalid pattern: /absolute/path"),
			expectedType:     "validation",
			expectedGraceful: true,
		},
		{
			name:             "generic error",
			err:              errors.New("unexpected error"),
			expectedType:     "generic",
			expectedGraceful: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorType := classifyError(tt.err)
			if errorType != tt.expectedType {
				t.Errorf("classifyError(%v) = %s, expected %s", tt.err, errorType, tt.expectedType)
			}

			graceful := isGracefulError(tt.err)
			if graceful != tt.expectedGraceful {
				t.Errorf("isGracefulError(%v) = %v, expected %v", tt.err, graceful, tt.expectedGraceful)
			}
		})
	}
}

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "network error",
			err:      &NetworkError{Message: "connection timeout"},
			expected: "Network error: connection timeout",
		},
		{
			name:     "registry error",
			err:      errors.New("registry 'test' not found"),
			expected: "Registry error: registry 'test' not found",
		},
		{
			name:     "generic error",
			err:      errors.New("unexpected error"),
			expected: "Error: unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorMessage(tt.err)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("formatErrorMessage(%v) should contain %q, got %q", tt.err, tt.expected, result)
			}
		})
	}
}

// classifyError determines the type of error for appropriate handling
func classifyError(err error) string {
	if err == nil {
		return "none"
	}

	errMsg := err.Error()

	if _, ok := err.(*NetworkError); ok {
		return "network"
	}

	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "registry") {
		return "registry"
	}

	if strings.Contains(errMsg, "invalid pattern") || strings.Contains(errMsg, "pattern cannot") {
		return "validation"
	}

	return "generic"
}

// isGracefulError determines if an error should be handled gracefully (not cause exit)
func isGracefulError(err error) bool {
	errorType := classifyError(err)
	return errorType == "network" || errorType == "registry" || errorType == "validation"
}

// formatErrorMessage formats error messages for user display
func formatErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errorType := classifyError(err)
	switch errorType {
	case "network":
		return "Network error: " + err.Error()
	case "registry":
		return "Registry error: " + err.Error()
	case "validation":
		return "Validation error: " + err.Error()
	default:
		return "Error: " + err.Error()
	}
}
