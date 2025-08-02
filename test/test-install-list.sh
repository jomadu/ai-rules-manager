#!/bin/bash

set -e

echo "=== ARM Install/List Test ==="

# Clean up any existing test files
rm -f rules.json rules.lock
rm -rf .cursorrules .amazonq

# Start test registry server in background
cd test/registry
go run server.go &
SERVER_PID=$!
cd ../..

# Wait for server to start
sleep 2

# Test 1: Install from manifest (should create default manifest)
echo "Test 1: Installing typescript-rules..."
go run ./cmd/arm install typescript-rules@1.0.0

# Test 2: List installed rulesets
echo "Test 2: Listing installed rulesets..."
go run ./cmd/arm list

# Test 3: Install another ruleset
echo "Test 3: Installing security-rules..."
go run ./cmd/arm install security-rules@1.2.0

# Test 4: List all rulesets
echo "Test 4: Listing all rulesets..."
go run ./cmd/arm list

# Test 5: List in JSON format
echo "Test 5: JSON format output..."
go run ./cmd/arm list --format=json

# Clean up
kill $SERVER_PID
echo "=== Test Complete ==="
