package errors

import (
	"fmt"
	"strings"
)

// ErrorCode represents a structured error code
type ErrorCode string

const (
	// Network errors
	ErrNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	ErrNetworkUnreachable ErrorCode = "NETWORK_UNREACHABLE"
	ErrPackageNotFound    ErrorCode = "PACKAGE_NOT_FOUND"
	ErrRegistryNotFound   ErrorCode = "REGISTRY_NOT_FOUND"

	// Configuration errors
	ErrConfigInvalid   ErrorCode = "CONFIG_INVALID"
	ErrConfigMissing   ErrorCode = "CONFIG_MISSING"
	ErrSourceNotFound  ErrorCode = "SOURCE_NOT_FOUND"
	ErrManifestInvalid ErrorCode = "MANIFEST_INVALID"

	// File system errors
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrDiskSpace        ErrorCode = "DISK_SPACE"
	ErrFileNotFound     ErrorCode = "FILE_NOT_FOUND"

	// Version errors
	ErrVersionInvalid  ErrorCode = "VERSION_INVALID"
	ErrVersionNotFound ErrorCode = "VERSION_NOT_FOUND"
	ErrVersionConflict ErrorCode = "VERSION_CONFLICT"
)

// ARMError represents a structured error with code, message, and suggestions
type ARMError struct {
	Code        ErrorCode
	Message     string
	Suggestions []string
	Context     map[string]string
	Cause       error
}

// Error implements the error interface
func (e *ARMError) Error() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("[%s] %s", e.Code, e.Message))

	if len(e.Context) > 0 {
		var contextParts []string
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
		}
		parts = append(parts, fmt.Sprintf("Context: %s", strings.Join(contextParts, ", ")))
	}

	if len(e.Suggestions) > 0 {
		parts = append(parts, fmt.Sprintf("Suggestions:\n  - %s", strings.Join(e.Suggestions, "\n  - ")))
	}

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Caused by: %v", e.Cause))
	}

	return strings.Join(parts, "\n")
}

// Unwrap returns the underlying cause for error wrapping
func (e *ARMError) Unwrap() error {
	return e.Cause
}

// New creates a new ARMError
func New(code ErrorCode, message string) *ARMError {
	return &ARMError{
		Code:        code,
		Message:     message,
		Context:     make(map[string]string),
		Suggestions: []string{},
	}
}

// Wrap wraps an existing error with ARM error context
func Wrap(err error, code ErrorCode, message string) *ARMError {
	return &ARMError{
		Code:        code,
		Message:     message,
		Context:     make(map[string]string),
		Suggestions: []string{},
		Cause:       err,
	}
}

// WithContext adds context to an error
func (e *ARMError) WithContext(key, value string) *ARMError {
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion to an error
func (e *ARMError) WithSuggestion(suggestion string) *ARMError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// Common error constructors
func NetworkTimeout(registry string) *ARMError {
	return New(ErrNetworkTimeout, fmt.Sprintf("Network timeout connecting to registry '%s'", registry)).
		WithContext("registry", registry).
		WithSuggestion("Check your internet connection").
		WithSuggestion("Verify registry URL with: arm config get sources." + registry)
}

func PackageNotFound(pkg, version, registry string) *ARMError {
	return New(ErrPackageNotFound, fmt.Sprintf("Package '%s@%s' not found in registry '%s'", pkg, version, registry)).
		WithContext("package", pkg).
		WithContext("version", version).
		WithContext("registry", registry).
		WithSuggestion("Check package name spelling").
		WithSuggestion("List available packages with: arm list --registry=" + registry).
		WithSuggestion("Try a different version or use version ranges like '^1.0.0'")
}

func ConfigInvalid(file, details string) *ARMError {
	return New(ErrConfigInvalid, fmt.Sprintf("Invalid configuration in '%s': %s", file, details)).
		WithContext("file", file).
		WithSuggestion("Check configuration syntax").
		WithSuggestion("See example config: arm config list")
}

func SourceNotFound(source string) *ARMError {
	return New(ErrSourceNotFound, fmt.Sprintf("Registry source '%s' not found in configuration", source)).
		WithContext("source", source).
		WithSuggestion("Add source with: arm config set sources." + source + " <URL>").
		WithSuggestion("List available sources with: arm config list")
}

func PermissionDenied(path string) *ARMError {
	return New(ErrPermissionDenied, fmt.Sprintf("Permission denied accessing '%s'", path)).
		WithContext("path", path).
		WithSuggestion("Check file/directory permissions").
		WithSuggestion("Try running with appropriate permissions")
}
