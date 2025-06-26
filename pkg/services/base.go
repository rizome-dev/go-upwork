// Package services provides service clients for the Upwork API.
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"github.com/rizome-dev/go-upwork/pkg/errors"
)

// BaseClient provides common functionality for all service clients
type BaseClient struct {
	HTTPClient     *http.Client
	APIURL         string
	OrganizationID string
	RateLimiter    RateLimiter
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage        `json:"data,omitempty"`
	Errors []errors.GraphQLError  `json:"errors,omitempty"`
}

// Do executes a GraphQL request
func (c *BaseClient) Do(ctx context.Context, req *GraphQLRequest, result interface{}) error {
	// Rate limiting
	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return err
		}
	}
	
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return errors.WrapError(err, "failed to marshal request")
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, bytes.NewReader(body))
	if err != nil {
		return errors.WrapError(err, "failed to create request")
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	
	if c.OrganizationID != "" {
		httpReq.Header.Set("X-Upwork-API-TenantId", c.OrganizationID)
	}
	
	// Execute request with retry
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = c.HTTPClient.Do(httpReq)
		if err != nil {
			if attempt < 2 && isRetryableError(err) {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return errors.WrapError(err, "request failed")
		}
		break
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WrapError(err, "failed to read response")
	}
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return c.handleHTTPError(resp.StatusCode, respBody)
	}
	
	// Parse GraphQL response
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &graphqlResp); err != nil {
		return errors.WrapError(err, "failed to parse response")
	}
	
	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return &errors.GraphQLErrors{Errors: graphqlResp.Errors}
	}
	
	// Unmarshal data if result is provided
	if result != nil && graphqlResp.Data != nil {
		if err := json.Unmarshal(graphqlResp.Data, result); err != nil {
			return errors.WrapError(err, "failed to unmarshal response data")
		}
	}
	
	return nil
}

// DoBatch executes multiple GraphQL requests in a single HTTP request
func (c *BaseClient) DoBatch(ctx context.Context, requests []*GraphQLRequest, results []interface{}) error {
	if len(requests) != len(results) {
		return fmt.Errorf("requests and results arrays must have the same length")
	}
	
	// Rate limiting
	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return err
		}
	}
	
	// Marshal batch request
	body, err := json.Marshal(requests)
	if err != nil {
		return errors.WrapError(err, "failed to marshal batch request")
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, bytes.NewReader(body))
	if err != nil {
		return errors.WrapError(err, "failed to create request")
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	
	if c.OrganizationID != "" {
		httpReq.Header.Set("X-Upwork-API-TenantId", c.OrganizationID)
	}
	
	// Execute request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return errors.WrapError(err, "batch request failed")
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WrapError(err, "failed to read response")
	}
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return c.handleHTTPError(resp.StatusCode, respBody)
	}
	
	// Parse batch response
	var batchResp []GraphQLResponse
	if err := json.Unmarshal(respBody, &batchResp); err != nil {
		return errors.WrapError(err, "failed to parse batch response")
	}
	
	// Process each response
	for i, graphqlResp := range batchResp {
		// Check for GraphQL errors
		if len(graphqlResp.Errors) > 0 {
			return fmt.Errorf("request %d failed: %v", i, &errors.GraphQLErrors{Errors: graphqlResp.Errors})
		}
		
		// Unmarshal data if result is provided
		if results[i] != nil && graphqlResp.Data != nil {
			if err := json.Unmarshal(graphqlResp.Data, results[i]); err != nil {
				return errors.WrapError(err, fmt.Sprintf("failed to unmarshal response %d", i))
			}
		}
	}
	
	return nil
}

// handleHTTPError handles HTTP error responses
func (c *BaseClient) handleHTTPError(statusCode int, body []byte) error {
	apiErr := &errors.APIError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
	}
	
	// Try to parse error response
	var errResp struct {
		Error   string                 `json:"error"`
		Message string                 `json:"message"`
		Code    string                 `json:"code"`
		Details map[string]interface{} `json:"details"`
	}
	
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Message != "" {
			apiErr.Message = errResp.Message
		} else if errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
		apiErr.Code = errResp.Code
		apiErr.Details = errResp.Details
	}
	
	return apiErr
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	// Implement retry logic for specific errors
	return errors.IsRetryable(err)
}

