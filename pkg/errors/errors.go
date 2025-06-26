// Package errors provides custom error types for the Upwork SDK.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	// Authentication errors
	ErrMissingCredentials = errors.New("missing client credentials")
	ErrNoRefreshToken     = errors.New("no refresh token available")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	
	// Request errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrRequestTimeout    = errors.New("request timeout")
	ErrInvalidRequest    = errors.New("invalid request")
	
	// API errors
	ErrNotFound          = errors.New("resource not found")
	ErrInternalServer    = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	
	// GraphQL errors
	ErrGraphQLParse      = errors.New("GraphQL parse error")
	ErrGraphQLValidation = errors.New("GraphQL validation error")
	ErrGraphQLExecution  = errors.New("GraphQL execution error")
)

// APIError represents an error returned by the Upwork API
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// Error returns the error message
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("upwork api error: %s - %s (status: %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("upwork api error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsNotFound returns true if the error is a not found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is an unauthorized error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsRateLimited returns true if the error is a rate limit error
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []Location             `json:"locations,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Location represents a location in the GraphQL query
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Error returns the error message
func (e *GraphQLError) Error() string {
	return e.Message
}

// GraphQLErrors represents multiple GraphQL errors
type GraphQLErrors struct {
	Errors []GraphQLError `json:"errors"`
}

// Error returns a combined error message
func (e *GraphQLErrors) Error() string {
	if len(e.Errors) == 0 {
		return "unknown GraphQL error"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple GraphQL errors: %s (and %d more)", e.Errors[0].Error(), len(e.Errors)-1)
}

// HasError returns true if there are any errors
func (e *GraphQLErrors) HasError() bool {
	return len(e.Errors) > 0
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error returns the error message
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for specific retryable errors
	if errors.Is(err, ErrRequestTimeout) ||
		errors.Is(err, ErrServiceUnavailable) ||
		errors.Is(err, ErrRateLimitExceeded) {
		return true
	}
	
	// Check for API errors
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500 || apiErr.StatusCode == http.StatusTooManyRequests
	}
	
	return false
}