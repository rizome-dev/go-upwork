package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rizome-dev/go-upwork/tests/mocks"
	"github.com/rizome-dev/go-upwork/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "valid endpoint",
			endpoint: "https://api.upwork.com/graphql",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "invalid endpoint",
			endpoint: "not-a-url",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.endpoint, nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.endpoint, client.endpoint)
			}
		})
	}
}

func TestSetHeader(t *testing.T) {
	client, err := NewClient("https://api.upwork.com/graphql", nil)
	require.NoError(t, err)

	client.SetHeader("X-Custom-Header", "custom-value")
	assert.Equal(t, "custom-value", client.headers["X-Custom-Header"])

	client.SetHeader("Authorization", "Bearer token")
	assert.Equal(t, "Bearer token", client.headers["Authorization"])
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		variables    map[string]interface{}
		mockResponse *http.Response
		mockError    error
		wantErr      bool
		validateReq  func(t *testing.T, req *http.Request)
		validateResp func(t *testing.T, resp map[string]interface{})
	}{
		{
			name:  "successful query",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			variables: map[string]interface{}{
				"id": "123",
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":   "123",
						"name": "John Doe",
					},
				},
				nil,
			)),
			wantErr: false,
			validateReq: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				
				var body map[string]interface{}
				err := json.NewDecoder(req.Body).Decode(&body)
				require.NoError(t, err)
				
				assert.Equal(t, `query GetUser($id: ID!) { user(id: $id) { id name } }`, body["query"])
				assert.Equal(t, map[string]interface{}{"id": "123"}, body["variables"])
			},
			validateResp: func(t *testing.T, resp map[string]interface{}) {
				user := resp["user"].(map[string]interface{})
				assert.Equal(t, "123", user["id"])
				assert.Equal(t, "John Doe", user["name"])
			},
		},
		{
			name:  "query with GraphQL errors",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			variables: map[string]interface{}{
				"id": "invalid",
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				nil,
				[]interface{}{
					testutils.CreateGraphQLError("User not found", "NOT_FOUND"),
				},
			)),
			wantErr: true,
		},
		{
			name:  "HTTP error",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockResponse: testutils.MockHTTPResponse(500, map[string]interface{}{
				"error": "Internal Server Error",
			}),
			wantErr: true,
		},
		{
			name:      "network error",
			query:     `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockError: fmt.Errorf("network error"),
			wantErr:   true,
		},
		{
			name:  "invalid JSON response",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("invalid json")),
				Header:     make(http.Header),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder()
			if tt.mockResponse != nil {
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       readResponseBody(t, tt.mockResponse),
				})
			}
			if tt.mockError != nil {
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					Error: tt.mockError,
				})
			}

			client, err := NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			var result map[string]interface{}
			err = client.Query(context.Background(), tt.query, tt.variables, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateResp != nil {
					tt.validateResp(t, result)
				}
			}

			if tt.validateReq != nil && len(recorder.Requests) > 0 {
				tt.validateReq(t, recorder.Requests[0])
			}
		})
	}
}

func TestMutate(t *testing.T) {
	tests := []struct {
		name         string
		mutation     string
		variables    map[string]interface{}
		mockResponse *http.Response
		wantErr      bool
		validateResp func(t *testing.T, resp map[string]interface{})
	}{
		{
			name:     "successful mutation",
			mutation: `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name } }`,
			variables: map[string]interface{}{
				"input": map[string]interface{}{
					"name":  "Jane Doe",
					"email": "jane@example.com",
				},
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				map[string]interface{}{
					"createUser": map[string]interface{}{
						"id":   "456",
						"name": "Jane Doe",
					},
				},
				nil,
			)),
			wantErr: false,
			validateResp: func(t *testing.T, resp map[string]interface{}) {
				user := resp["createUser"].(map[string]interface{})
				assert.Equal(t, "456", user["id"])
				assert.Equal(t, "Jane Doe", user["name"])
			},
		},
		{
			name:     "mutation with validation errors",
			mutation: `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name } }`,
			variables: map[string]interface{}{
				"input": map[string]interface{}{
					"name": "", // Invalid empty name
				},
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				nil,
				[]interface{}{
					testutils.CreateGraphQLError("Name is required", "VALIDATION_ERROR"),
				},
			)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder()
			if tt.mockResponse != nil {
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       readResponseBody(t, tt.mockResponse),
				})
			}

			client, err := NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			var result map[string]interface{}
			err = client.Mutate(context.Background(), tt.mutation, tt.variables, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateResp != nil {
					tt.validateResp(t, result)
				}
			}
		})
	}
}

func TestRawRequest(t *testing.T) {
	tests := []struct {
		name         string
		request      Request
		mockResponse *http.Response
		wantErr      bool
		validateReq  func(t *testing.T, req *http.Request)
	}{
		{
			name: "raw request with operation name",
			request: Request{
				Query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
				Variables: map[string]interface{}{
					"id": "123",
				},
				OperationName: "GetUser",
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":   "123",
						"name": "John Doe",
					},
				},
				nil,
			)),
			wantErr: false,
			validateReq: func(t *testing.T, req *http.Request) {
				var body map[string]interface{}
				err := json.NewDecoder(req.Body).Decode(&body)
				require.NoError(t, err)
				
				assert.Equal(t, "GetUser", body["operationName"])
			},
		},
		{
			name: "raw request with custom headers",
			request: Request{
				Query: `{ viewer { id } }`,
			},
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				map[string]interface{}{
					"viewer": map[string]interface{}{
						"id": "viewer123",
					},
				},
				nil,
			)),
			wantErr: false,
			validateReq: func(t *testing.T, req *http.Request) {
				// Check that custom headers are preserved
				assert.Equal(t, "custom-value", req.Header.Get("X-Custom-Header"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder()
			if tt.mockResponse != nil {
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       readResponseBody(t, tt.mockResponse),
				})
			}

			client, err := NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)
			
			// Set a custom header for testing
			client.SetHeader("X-Custom-Header", "custom-value")

			response, err := client.RawRequest(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}

			if tt.validateReq != nil && len(recorder.Requests) > 0 {
				tt.validateReq(t, recorder.Requests[0])
			}
		})
	}
}

func TestBatchRequest(t *testing.T) {
	recorder := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body: `[
				{
					"data": {
						"user": {
							"id": "123",
							"name": "John Doe"
						}
					}
				},
				{
					"data": {
						"contracts": [
							{"id": "c1", "title": "Contract 1"},
							{"id": "c2", "title": "Contract 2"}
						]
					}
				}
			]`,
		},
	)

	client, err := NewClient("https://api.upwork.com/graphql", recorder)
	require.NoError(t, err)

	requests := []Request{
		{
			Query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			Variables: map[string]interface{}{
				"id": "123",
			},
		},
		{
			Query: `query GetContracts { contracts { id title } }`,
		},
	}

	responses, err := client.BatchRequest(context.Background(), requests)
	require.NoError(t, err)
	assert.Len(t, responses, 2)

	// Verify first response
	userData, ok := responses[0].Data.(map[string]interface{})
	require.True(t, ok)
	user := userData["user"].(map[string]interface{})
	assert.Equal(t, "123", user["id"])
	assert.Equal(t, "John Doe", user["name"])

	// Verify second response
	contractData, ok := responses[1].Data.(map[string]interface{})
	require.True(t, ok)
	contracts := contractData["contracts"].([]interface{})
	assert.Len(t, contracts, 2)
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse *http.Response
		expectedErr  string
	}{
		{
			name: "401 Unauthorized",
			mockResponse: testutils.MockHTTPResponse(401, map[string]interface{}{
				"error": "Unauthorized",
			}),
			expectedErr: "401",
		},
		{
			name: "403 Forbidden",
			mockResponse: testutils.MockHTTPResponse(403, map[string]interface{}{
				"error": "Forbidden",
			}),
			expectedErr: "403",
		},
		{
			name: "429 Rate Limited",
			mockResponse: testutils.MockHTTPResponse(429, map[string]interface{}{
				"error": "Too Many Requests",
			}),
			expectedErr: "429",
		},
		{
			name: "500 Internal Server Error",
			mockResponse: testutils.MockHTTPResponse(500, map[string]interface{}{
				"error": "Internal Server Error",
			}),
			expectedErr: "500",
		},
		{
			name: "GraphQL field errors",
			mockResponse: testutils.MockHTTPResponse(200, map[string]interface{}{
				"errors": []interface{}{
					map[string]interface{}{
						"message": "Cannot query field 'invalid' on type 'User'",
						"locations": []interface{}{
							map[string]interface{}{
								"line":   2,
								"column": 5,
							},
						},
						"path": []interface{}{"user", "invalid"},
					},
				},
			}),
			expectedErr: "Cannot query field 'invalid' on type 'User'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder(
				mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       readResponseBody(t, tt.mockResponse),
				},
			)

			client, err := NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			var result map[string]interface{}
			err = client.Query(context.Background(), `{ user { id } }`, nil, &result)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a recorder that simulates a slow response
	recorder := &mocks.RequestRecorder{}
	recorder.Do = func(req *http.Request) (*http.Response, error) {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-req.Context().Done():
			return nil, fmt.Errorf("request cancelled")
		}
	}

	client, err := NewClient("https://api.upwork.com/graphql", recorder)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result map[string]interface{}
	err = client.Query(ctx, `{ user { id } }`, nil, &result)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestHeaderPropagation(t *testing.T) {
	recorder := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       `{"data": {"test": "ok"}}`,
		},
	)

	client, err := NewClient("https://api.upwork.com/graphql", recorder)
	require.NoError(t, err)

	// Set various headers
	client.SetHeader("Authorization", "Bearer test-token")
	client.SetHeader("X-Organization-ID", "org-123")
	client.SetHeader("X-Request-ID", "req-456")

	var result map[string]interface{}
	err = client.Query(context.Background(), `{ test }`, nil, &result)
	require.NoError(t, err)

	// Verify headers were sent
	req := recorder.GetLastRequest()
	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	assert.Equal(t, "org-123", req.Header.Get("X-Organization-ID"))
	assert.Equal(t, "req-456", req.Header.Get("X-Request-ID"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

// Helper function to read response body
func readResponseBody(t *testing.T, resp *http.Response) string {
	if resp.Body == nil {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}