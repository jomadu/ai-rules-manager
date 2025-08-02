package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
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

	// Expected format: /package/version.tar.gz
	packagePath := filepath.Join("packages", r.URL.Path)

	http.ServeFile(w, r, packagePath)
}
