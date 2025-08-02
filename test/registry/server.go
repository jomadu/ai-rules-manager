package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	http.HandleFunc("/", handlePackageRequest)

	fmt.Println("Test registry server starting on :8080")
	fmt.Println("Available packages:")
	fmt.Println("  - typescript-rules@1.0.0")
	fmt.Println("  - security-rules@1.2.0")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePackageRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

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

	packagePath := filepath.Join("packages", cleanPath)
	http.ServeFile(w, r, packagePath)
}
