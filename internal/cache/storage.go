package cache

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct {
	basePath string
}

type CacheConfig struct {
	PackageTTL  time.Duration
	MetadataTTL time.Duration
	VersionTTL  time.Duration
}

var DefaultConfig = CacheConfig{
	PackageTTL:  0, // Never expire
	MetadataTTL: time.Hour,
	VersionTTL:  15 * time.Minute,
}

func New() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".arm", "cache")
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Storage{basePath: basePath}, nil
}

func (s *Storage) PackagePath(registryURL, ruleset, version string) string {
	host := extractHost(registryURL)
	return filepath.Join(s.basePath, "packages", host, ruleset, version)
}

func (s *Storage) MetadataPath(registryURL string) string {
	host := extractHost(registryURL)
	return filepath.Join(s.basePath, "registry", host)
}

func extractHost(registryURL string) string {
	if registryURL == "" {
		return "unknown"
	}

	parsed, err := url.Parse(registryURL)
	if err != nil {
		// Fallback: hash the URL if parsing fails
		hash := sha256.Sum256([]byte(registryURL))
		return fmt.Sprintf("hash-%x", hash[:8])
	}

	host := parsed.Host
	if host == "" {
		// Handle file:// URLs and invalid URLs
		if strings.HasPrefix(registryURL, "file://") {
			host = strings.TrimPrefix(registryURL, "file://")
			host = strings.ReplaceAll(host, string(filepath.Separator), "_")
		} else {
			// For URLs without host (like "invalid-url"), hash them
			hash := sha256.Sum256([]byte(registryURL))
			host = fmt.Sprintf("hash-%x", hash[:8])
		}
	}

	return host
}

func (s *Storage) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Storage) IsExpired(path string, ttl time.Duration) bool {
	if ttl == 0 {
		return false // Never expires
	}

	info, err := os.Stat(path)
	if err != nil {
		return true // Doesn't exist = expired
	}

	return time.Since(info.ModTime()) > ttl
}
