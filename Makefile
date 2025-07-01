# Makefile for Upwork Go SDK

.PHONY: help test test-coverage test-unit test-integration test-race test-bench lint clean docs

# Default target
help:
	@echo "Upwork Go SDK - Available targets:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-race      - Run tests with race detector"
	@echo "  make test-bench     - Run benchmarks"
	@echo "  make lint          - Run linter"
	@echo "  make clean         - Clean build and test artifacts"
	@echo "  make docs          - Generate documentation"

# Run all tests
test:
	@echo "Running all tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	@go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated at coverage/coverage.html"
	@go tool cover -func=coverage/coverage.out | grep -E '^total:'

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race -short ./...

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Clean build and test artifacts
clean:
	@echo "Cleaning artifacts..."
	@rm -rf coverage/
	@rm -f *.out
	@rm -f *.test
	@go clean -testcache

# Generate documentation
docs:
	@echo "Generating documentation..."
	@go doc -all ./pkg > docs/API.md
	@echo "API documentation generated at docs/API.md"

# Quick test for CI
ci-test: lint test-race test-coverage
	@echo "CI tests completed"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Check for security vulnerabilities
security:
	@echo "Checking for vulnerabilities..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

# Run a specific test
test-specific:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific TEST=TestName"; \
		exit 1; \
	fi
	@echo "Running test: $(TEST)"
	@go test -v -run $(TEST) ./...

# Test a specific package
test-package:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-package PKG=./pkg/services"; \
		exit 1; \
	fi
	@echo "Testing package: $(PKG)"
	@go test -v $(PKG)/...