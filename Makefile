.PHONY: build test lint fmt clean install-tools setup-hooks

# Build the binary
build:
	go build -o arm ./cmd/arm

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	$(shell go env GOPATH)/bin/golangci-lint run

# Format code
fmt:
	gofmt -w .
	$(shell go env GOPATH)/bin/goimports -w .

# Clean build artifacts
clean:
	rm -f arm coverage.out
	rm -rf .venv

# Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup pre-commit hooks with virtual environment
setup-hooks:
	python3 -m venv .venv
	.venv/bin/pip install pre-commit
	.venv/bin/pre-commit install
	.venv/bin/pre-commit install --hook-type commit-msg

# Run all checks
check: fmt lint test

# Development setup
setup: install-tools setup-hooks
	go mod tidy
