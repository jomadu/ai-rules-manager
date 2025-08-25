package errors

import "fmt"

// RegistryError represents registry-related errors
type RegistryError struct {
	Registry string
	Op       string
	Err      error
}

func (e *RegistryError) Error() string {
	return fmt.Sprintf("registry %s: %s: %v", e.Registry, e.Op, e.Err)
}

func (e *RegistryError) Unwrap() error {
	return e.Err
}

// CacheError represents cache-related errors
type CacheError struct {
	Key string
	Op  string
	Err error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s: %s: %v", e.Key, e.Op, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	File string
	Op   string
	Err  error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config %s: %s: %v", e.File, e.Op, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// VersionError represents version resolution errors
type VersionError struct {
	Constraint string
	Op         string
	Err        error
}

func (e *VersionError) Error() string {
	return fmt.Sprintf("version %s: %s: %v", e.Constraint, e.Op, e.Err)
}

func (e *VersionError) Unwrap() error {
	return e.Err
}
