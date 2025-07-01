#!/bin/bash

# Script to run comprehensive tests for the Upwork Go SDK

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running Upwork Go SDK Test Suite${NC}"
echo "======================================"

# Create coverage directory if it doesn't exist
mkdir -p coverage

# Run tests with coverage
echo -e "\n${YELLOW}Running unit tests with coverage...${NC}"
go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...

# Generate coverage report
echo -e "\n${YELLOW}Generating coverage report...${NC}"
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Show coverage summary
echo -e "\n${YELLOW}Coverage Summary:${NC}"
go tool cover -func=coverage/coverage.out | grep -E '^total:|^github.com/rizome-dev/go-upwork/pkg|^github.com/rizome-dev/go-upwork/internal'

# Run specific test suites
echo -e "\n${YELLOW}Running test suites by category:${NC}"

echo -e "\n${GREEN}1. Authentication Tests${NC}"
go test -v ./pkg/auth/... -count=1

echo -e "\n${GREEN}2. GraphQL Client Tests${NC}"
go test -v ./internal/graphql/... -count=1

echo -e "\n${GREEN}3. Service Layer Tests${NC}"
go test -v ./pkg/services/... -count=1

echo -e "\n${GREEN}4. Main Client Tests${NC}"
go test -v ./pkg/client_test.go -count=1

# Run benchmarks if requested
if [[ "$1" == "--bench" ]]; then
    echo -e "\n${YELLOW}Running benchmarks...${NC}"
    go test -bench=. -benchmem ./...
fi

# Run with race detector
echo -e "\n${YELLOW}Running race condition tests...${NC}"
go test -race -short ./...

# Check for test failures
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}"
    echo -e "Coverage report available at: coverage/coverage.html"
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    exit 1
fi

# Optional: Run linter if golangci-lint is installed
if command -v golangci-lint &> /dev/null; then
    echo -e "\n${YELLOW}Running linter...${NC}"
    golangci-lint run ./...
fi