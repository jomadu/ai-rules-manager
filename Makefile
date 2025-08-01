.PHONY: build test lint fmt clean install-tools setup-hooks

# Build the binary
build:
	go build -o arm ./cmd/arm

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -w .
	goimports -w .

# Clean build artifacts
clean:
	rm -f arm coverage.out

# Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Setup pre-commit hooks
setup-hooks:
	pip install pre-commit
	pre-commit install
	pre-commit install --hook-type commit-msg

# Run all checks
check: fmt lint test

# Development setup
setup: install-tools setup-hooks
	go mod tidy