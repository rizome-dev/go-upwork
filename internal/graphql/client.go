// Package graphql provides internal GraphQL client functionality.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Request represents a GraphQL request
type Request struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// Response represents a GraphQL response
type Response struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []Error         `json:"errors,omitempty"`
}

// Error represents a GraphQL error
type Error struct {
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

// Client is an internal GraphQL client
type Client struct {
	httpClient *http.Client
	endpoint   string
	headers    map[string]string
}

// NewClient creates a new GraphQL client
func NewClient(httpClient *http.Client, endpoint string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		headers:    make(map[string]string),
	}
}

// SetHeader sets a header for all requests
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// Do executes a GraphQL request
func (c *Client) Do(ctx context.Context, req *Request, result interface{}) error {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse GraphQL response
	var graphqlResp Response
	if err := json.Unmarshal(respBody, &graphqlResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return &ErrorList{Errors: graphqlResp.Errors}
	}
	
	// Unmarshal data if result is provided
	if result != nil && graphqlResp.Data != nil {
		if err := json.Unmarshal(graphqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal response data: %w", err)
		}
	}
	
	return nil
}

// ErrorList represents multiple GraphQL errors
type ErrorList struct {
	Errors []Error
}

// Error returns a combined error message
func (e *ErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "unknown GraphQL error"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Message
	}
	return fmt.Sprintf("multiple GraphQL errors: %s (and %d more)", e.Errors[0].Message, len(e.Errors)-1)
}

// HasError returns true if there are any errors
func (e *ErrorList) HasError() bool {
	return len(e.Errors) > 0
}