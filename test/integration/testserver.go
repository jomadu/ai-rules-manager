package integration

import (
	"net/http"
	"path/filepath"
	"strings"
)

// NewTestServer creates a test HTTP server for registry testing
func NewTestServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlePackageRequest)
	return mux
}

func handlePackageRequest(w http.ResponseWriter, r *http.Request) {
	// Sanitize path to prevent directory traversal
	cleanPath := filepath.Clean(r.URL.Path)
	if cleanPath == "." || cleanPath == "/" {
		http.NotFound(w, r)
		return
	}

	// Remove leading slash and validate path
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	if strings.Contains(cleanPath, "..") {
		http.NotFound(w, r)
		return
	}

	// For integration tests, serve from test/registry/packages
	baseDir := filepath.Join("..", "..", "test", "registry", "packages")
	packagePath := filepath.Join(baseDir, cleanPath)
	http.ServeFile(w, r, packagePath)
}
