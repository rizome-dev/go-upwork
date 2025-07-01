# Upwork Go SDK Testing Guide

## Overview

This document provides comprehensive information about the test suite for the Upwork Go SDK. The test suite ensures reliability, correctness, and maintainability of the SDK through extensive unit tests, integration tests, and mocking strategies.

## Test Structure

```
tests/
├── mocks/           # Mock implementations for testing
│   ├── http_client.go
│   └── rate_limiter.go
├── testutils/       # Testing utilities and helpers
│   └── testutils.go
pkg/
├── auth/
│   └── oauth2_test.go
├── services/
│   ├── base_test.go
│   └── contracts_test.go
├── client_test.go
internal/
└── graphql/
    └── client_test.go
```

## Running Tests

### Run All Tests
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests Using the Test Script
```bash
./scripts/test.sh
```

### Run Specific Test Suites
```bash
# Authentication tests
go test -v ./pkg/auth/...

# GraphQL client tests
go test -v ./internal/graphql/...

# Service layer tests
go test -v ./pkg/services/...

# Main client tests
go test -v ./pkg/...
```

### Run with Race Detection
```bash
go test -race ./...
```

### Run Benchmarks
```bash
./scripts/test.sh --bench
```

## Test Categories

### 1. Authentication Tests (`pkg/auth/oauth2_test.go`)

Tests OAuth2 authentication flows including:
- OAuth2 configuration validation
- Authorization URL generation
- Authorization code exchange
- Token refresh
- Client credentials grant
- Token validation
- Error handling
- Concurrent token refresh

**Key Test Cases:**
- Valid and invalid configurations
- Token lifecycle management
- OAuth2 error responses
- Concurrent access patterns

### 2. GraphQL Client Tests (`internal/graphql/client_test.go`)

Tests the GraphQL client implementation:
- Client initialization
- Query execution
- Mutation execution
- Batch requests
- Error handling
- Context cancellation
- Header propagation

**Key Test Cases:**
- Successful queries and mutations
- GraphQL error handling
- HTTP error handling
- Rate limiting
- Request/response serialization

### 3. Base Service Tests (`pkg/services/base_test.go`)

Tests the common service functionality:
- Query and mutation execution
- Retry logic
- Rate limiting integration
- Error wrapping
- Pagination support
- Batch requests

**Key Test Cases:**
- Retry behavior for transient errors
- Rate limit handling
- Context cancellation
- Error type mapping

### 4. Service Implementation Tests

Example: Contracts Service (`pkg/services/contracts_test.go`)
- CRUD operations
- Business logic validation
- GraphQL query construction
- Response parsing
- Error scenarios

**Key Test Cases:**
- Get contract by ID
- List contracts with filters
- Create/update contracts
- Contract lifecycle operations

### 5. Main Client Tests (`pkg/client_test.go`)

Tests the main SDK client:
- Client initialization
- Service initialization
- Token management
- Configuration options
- Thread safety

**Key Test Cases:**
- Configuration validation
- Service availability
- Concurrent token access
- Organization ID management

## Mock Implementations

### HTTP Client Mock (`tests/mocks/http_client.go`)

Provides mock HTTP clients for testing:
- `MockHTTPClient`: Simple response mocking
- `RequestRecorder`: Records requests and provides responses
- `HTTPRoundTripper`: Mock implementation of `http.RoundTripper`

### Rate Limiter Mock (`tests/mocks/rate_limiter.go`)

Provides mock rate limiters:
- `MockRateLimiter`: Configurable rate limiting behavior
- `NoOpRateLimiter`: No-op implementation
- `RecordingRateLimiter`: Records rate limit calls

## Test Utilities (`tests/testutils/testutils.go`)

Common utilities for testing:
- `MockHTTPResponse`: Creates mock HTTP responses
- `MockGraphQLResponse`: Creates GraphQL responses
- `MockOAuth2Token`: Creates OAuth2 token responses
- `CreateGraphQLError`: Creates GraphQL errors
- Sample data generators for various API objects

## Writing New Tests

### 1. Unit Test Template

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        want     interface{}
        wantErr  bool
    }{
        {
            name:    "successful case",
            input:   "valid input",
            want:    "expected output",
            wantErr: false,
        },
        {
            name:    "error case",
            input:   "invalid input",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Service Test Template

```go
func setupService(responses ...mocks.MockResponse) (*ServiceType, *mocks.RequestRecorder) {
    recorder := mocks.NewRequestRecorder(responses...)
    client, _ := graphql.NewClient("https://api.upwork.com/graphql", recorder)
    rateLimiter := mocks.NewMockRateLimiter()
    service := NewServiceType(client, rateLimiter)
    return service, recorder
}

func TestServiceMethod(t *testing.T) {
    mockResponse := testutils.MockGraphQLResponse(
        map[string]interface{}{
            "data": "response data",
        },
        nil,
    )
    
    body, _ := json.Marshal(mockResponse)
    service, recorder := setupService(
        mocks.MockResponse{
            StatusCode: 200,
            Body:       string(body),
        },
    )
    
    // Test the service method
    result, err := service.Method(context.Background(), params)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Verify the request
    req := recorder.GetLastRequest()
    // Additional request verification
}
```

## Coverage Goals

The test suite aims for:
- **Overall Coverage**: >80%
- **Critical Paths**: 100% (authentication, error handling)
- **Service Layer**: >90%
- **Edge Cases**: Comprehensive coverage

## Continuous Integration

Tests should be run in CI/CD pipelines with:
- All tests must pass
- Coverage threshold enforcement
- Race condition detection
- Linting checks

## Best Practices

1. **Use Table-Driven Tests**: Organize test cases in tables for clarity
2. **Mock External Dependencies**: Use mocks for HTTP clients and rate limiters
3. **Test Error Scenarios**: Always test both success and failure paths
4. **Verify Request Content**: Check that requests contain expected data
5. **Use Test Helpers**: Leverage test utilities for common operations
6. **Isolate Tests**: Each test should be independent
7. **Clear Test Names**: Use descriptive names that explain what is being tested
8. **Assert All Outputs**: Verify all return values, not just errors

## Debugging Tests

### Verbose Output
```bash
go test -v ./path/to/package
```

### Run Specific Test
```bash
go test -v -run TestSpecificFunction ./path/to/package
```

### Debug with Delve
```bash
dlv test ./path/to/package -- -test.run TestSpecificFunction
```

### View Coverage for Specific Package
```bash
go test -coverprofile=coverage.out ./path/to/package
go tool cover -func=coverage.out
```

## Integration Testing

While most tests use mocks, integration tests can be run against the actual Upwork API:

```go
// +build integration

func TestRealAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Use real credentials from environment
    clientID := os.Getenv("UPWORK_CLIENT_ID")
    clientSecret := os.Getenv("UPWORK_CLIENT_SECRET")
    
    // Test against real API
}
```

Run integration tests:
```bash
go test -tags=integration ./...
```

## Performance Testing

### Benchmarks

```go
func BenchmarkGraphQLQuery(b *testing.B) {
    service := setupTestService()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = service.Query(ctx, "query", nil)
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

## Troubleshooting

### Common Issues

1. **Mock Response Mismatch**: Ensure mock responses match expected GraphQL schema
2. **Race Conditions**: Use `-race` flag to detect concurrent access issues
3. **Context Cancellation**: Test timeout and cancellation scenarios
4. **Token Expiry**: Mock various token states (valid, expired, missing)

### Test Maintenance

- Update tests when API changes
- Keep mocks synchronized with actual API responses
- Review and update test data regularly
- Monitor test execution time

## Contributing

When contributing tests:
1. Follow existing patterns
2. Maintain high coverage
3. Document complex test scenarios
4. Update this guide for new test categories