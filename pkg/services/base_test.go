package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/rizome-dev/go-upwork/internal/graphql"
	"github.com/rizome-dev/go-upwork/pkg/models"
	upworkErrors "github.com/rizome-dev/go-upwork/pkg/errors"
	"github.com/rizome-dev/go-upwork/tests/mocks"
	"github.com/rizome-dev/go-upwork/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseService(t *testing.T) {
	mockClient, _ := graphql.NewClient("https://api.upwork.com/graphql", nil)
	rateLimiter := mocks.NewMockRateLimiter()

	service := NewBaseService(mockClient, rateLimiter)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockClient, service.client)
	assert.Equal(t, rateLimiter, service.rateLimiter)
}

func TestExecuteQuery(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		variables    map[string]interface{}
		mockResponse *http.Response
		mockError    error
		rateLimitErr error
		wantErr      bool
		expectedData map[string]interface{}
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
			expectedData: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John Doe",
				},
			},
		},
		{
			name:  "rate limit error",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			rateLimitErr: errors.New("rate limit exceeded"),
			wantErr: true,
		},
		{
			name:  "GraphQL error",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockResponse: testutils.MockHTTPResponse(200, testutils.MockGraphQLResponse(
				nil,
				[]interface{}{
					testutils.CreateGraphQLError("User not found", "NOT_FOUND"),
				},
			)),
			wantErr: true,
		},
		{
			name:  "HTTP error with retry",
			query: `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockResponse: testutils.MockHTTPResponse(500, map[string]interface{}{
				"error": "Internal Server Error",
			}),
			wantErr: true,
		},
		{
			name:      "network error",
			query:     `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			mockError: errors.New("network error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder()
			if tt.mockResponse != nil {
				body, _ := json.Marshal(tt.mockResponse.Body)
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       string(body),
				})
			}
			if tt.mockError != nil {
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					Error: tt.mockError,
				})
			}

			client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			rateLimiter := mocks.NewMockRateLimiter()
			if tt.rateLimitErr != nil {
				rateLimiter.ShouldError = true
				rateLimiter.Error = tt.rateLimitErr
			}

			service := NewBaseService(client, rateLimiter)

			var result map[string]interface{}
			err = service.ExecuteQuery(context.Background(), tt.query, tt.variables, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}

			// Verify rate limiter was called
			assert.Equal(t, 1, rateLimiter.GetWaitCalls())
		})
	}
}

func TestExecuteMutation(t *testing.T) {
	tests := []struct {
		name         string
		mutation     string
		variables    map[string]interface{}
		mockResponse *http.Response
		wantErr      bool
		expectedData map[string]interface{}
	}{
		{
			name:     "successful mutation",
			mutation: `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name } }`,
			variables: map[string]interface{}{
				"input": map[string]interface{}{
					"name": "Jane Doe",
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
			expectedData: map[string]interface{}{
				"createUser": map[string]interface{}{
					"id":   "456",
					"name": "Jane Doe",
				},
			},
		},
		{
			name:     "mutation with validation error",
			mutation: `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name } }`,
			variables: map[string]interface{}{
				"input": map[string]interface{}{},
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
				body, _ := json.Marshal(tt.mockResponse.Body)
				recorder.Responses = append(recorder.Responses, mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       string(body),
				})
			}

			client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			service := NewBaseService(client, mocks.NewMockRateLimiter())

			var result map[string]interface{}
			err = service.ExecuteMutation(context.Background(), tt.mutation, tt.variables, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestRetryLogic(t *testing.T) {
	tests := []struct {
		name          string
		responses     []mocks.MockResponse
		expectedCalls int
		wantErr       bool
	}{
		{
			name: "success on first try",
			responses: []mocks.MockResponse{
				{
					StatusCode: 200,
					Body:       `{"data": {"test": "ok"}}`,
				},
			},
			expectedCalls: 1,
			wantErr:       false,
		},
		{
			name: "success after retry",
			responses: []mocks.MockResponse{
				{
					StatusCode: 500,
					Body:       `{"error": "Internal Server Error"}`,
				},
				{
					StatusCode: 200,
					Body:       `{"data": {"test": "ok"}}`,
				},
			},
			expectedCalls: 2,
			wantErr:       false,
		},
		{
			name: "fail after max retries",
			responses: []mocks.MockResponse{
				{
					StatusCode: 500,
					Body:       `{"error": "Internal Server Error"}`,
				},
				{
					StatusCode: 500,
					Body:       `{"error": "Internal Server Error"}`,
				},
				{
					StatusCode: 500,
					Body:       `{"error": "Internal Server Error"}`,
				},
			},
			expectedCalls: 3,
			wantErr:       true,
		},
		{
			name: "non-retryable error",
			responses: []mocks.MockResponse{
				{
					StatusCode: 401,
					Body:       `{"error": "Unauthorized"}`,
				},
			},
			expectedCalls: 1,
			wantErr:       true,
		},
		{
			name: "rate limit error should retry",
			responses: []mocks.MockResponse{
				{
					StatusCode: 429,
					Body:       `{"error": "Rate limit exceeded"}`,
				},
				{
					StatusCode: 200,
					Body:       `{"data": {"test": "ok"}}`,
				},
			},
			expectedCalls: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := mocks.NewRequestRecorder(tt.responses...)
			
			client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			service := NewBaseService(client, mocks.NewMockRateLimiter())

			var result map[string]interface{}
			err = service.ExecuteQuery(context.Background(), `{ test }`, nil, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCalls, recorder.CallCount)
		})
	}
}

func TestBatchRequest(t *testing.T) {
	mockResponses := `[
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
	]`

	recorder := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       mockResponses,
		},
	)

	client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
	require.NoError(t, err)

	service := NewBaseService(client, mocks.NewMockRateLimiter())

	requests := []graphql.Request{
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

	responses, err := service.ExecuteBatchRequest(context.Background(), requests)
	require.NoError(t, err)
	assert.Len(t, responses, 2)

	// Verify first response
	userData, ok := responses[0].Data.(map[string]interface{})
	require.True(t, ok)
	user := userData["user"].(map[string]interface{})
	assert.Equal(t, "123", user["id"])

	// Verify second response
	contractData, ok := responses[1].Data.(map[string]interface{})
	require.True(t, ok)
	contracts := contractData["contracts"].([]interface{})
	assert.Len(t, contracts, 2)
}

func TestPaginatedQuery(t *testing.T) {
	// Mock responses for paginated queries
	firstPageResponse := testutils.MockGraphQLResponse(
		map[string]interface{}{
			"contracts": map[string]interface{}{
				"edges": []interface{}{
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":    "1",
							"title": "Contract 1",
						},
					},
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":    "2",
							"title": "Contract 2",
						},
					},
				},
				"pageInfo": map[string]interface{}{
					"hasNextPage": true,
					"endCursor":   "cursor2",
				},
			},
		},
		nil,
	)

	secondPageResponse := testutils.MockGraphQLResponse(
		map[string]interface{}{
			"contracts": map[string]interface{}{
				"edges": []interface{}{
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":    "3",
							"title": "Contract 3",
						},
					},
				},
				"pageInfo": map[string]interface{}{
					"hasNextPage": false,
					"endCursor":   "cursor3",
				},
			},
		},
		nil,
	)

	body1, _ := json.Marshal(firstPageResponse)
	body2, _ := json.Marshal(secondPageResponse)

	recorder := mocks.NewRequestRecorder(
		mocks.MockResponse{
			StatusCode: 200,
			Body:       string(body1),
		},
		mocks.MockResponse{
			StatusCode: 200,
			Body:       string(body2),
		},
	)

	client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
	require.NoError(t, err)

	service := NewBaseService(client, mocks.NewMockRateLimiter())

	// Test paginated query collection
	query := `
		query GetContracts($after: String) {
			contracts(first: 2, after: $after) {
				edges {
					node {
						id
						title
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	allContracts := []interface{}{}
	var pageInfo models.PageInfo
	cursor := ""

	for {
		var result struct {
			Contracts struct {
				Edges []struct {
					Node map[string]interface{} `json:"node"`
				} `json:"edges"`
				PageInfo models.PageInfo `json:"pageInfo"`
			} `json:"contracts"`
		}

		variables := map[string]interface{}{}
		if cursor != "" {
			variables["after"] = cursor
		}

		err := service.ExecuteQuery(context.Background(), query, variables, &result)
		require.NoError(t, err)

		for _, edge := range result.Contracts.Edges {
			allContracts = append(allContracts, edge.Node)
		}

		pageInfo = result.Contracts.PageInfo
		if !pageInfo.HasNextPage {
			break
		}
		cursor = pageInfo.EndCursor
	}

	assert.Len(t, allContracts, 3)
	assert.Equal(t, "1", allContracts[0].(map[string]interface{})["id"])
	assert.Equal(t, "2", allContracts[1].(map[string]interface{})["id"])
	assert.Equal(t, "3", allContracts[2].(map[string]interface{})["id"])
}

func TestContextCancellation(t *testing.T) {
	// Create a rate limiter that waits
	rateLimiter := mocks.NewMockRateLimiter()
	rateLimiter.WaitDuration = 100 * time.Millisecond

	client, err := graphql.NewClient("https://api.upwork.com/graphql", nil)
	require.NoError(t, err)

	service := NewBaseService(client, rateLimiter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result map[string]interface{}
	err = service.ExecuteQuery(ctx, `{ test }`, nil, &result)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *http.Response
		expectedError error
		checkType     func(error) bool
	}{
		{
			name: "rate limit error",
			mockResponse: testutils.MockHTTPResponse(429, map[string]interface{}{
				"error": "Rate limit exceeded",
			}),
			checkType: func(err error) bool {
				var rateLimitErr *upworkErrors.RateLimitError
				return errors.As(err, &rateLimitErr)
			},
		},
		{
			name: "authentication error",
			mockResponse: testutils.MockHTTPResponse(401, map[string]interface{}{
				"error": "Unauthorized",
			}),
			checkType: func(err error) bool {
				var authErr *upworkErrors.AuthenticationError
				return errors.As(err, &authErr)
			},
		},
		{
			name: "permission error",
			mockResponse: testutils.MockHTTPResponse(403, map[string]interface{}{
				"error": "Forbidden",
			}),
			checkType: func(err error) bool {
				var permErr *upworkErrors.PermissionError
				return errors.As(err, &permErr)
			},
		},
		{
			name: "validation error",
			mockResponse: testutils.MockHTTPResponse(400, map[string]interface{}{
				"error": "Bad Request",
			}),
			checkType: func(err error) bool {
				var valErr *upworkErrors.ValidationError
				return errors.As(err, &valErr)
			},
		},
		{
			name: "server error",
			mockResponse: testutils.MockHTTPResponse(500, map[string]interface{}{
				"error": "Internal Server Error",
			}),
			checkType: func(err error) bool {
				var serverErr *upworkErrors.ServerError
				return errors.As(err, &serverErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.mockResponse.Body)
			recorder := mocks.NewRequestRecorder(
				mocks.MockResponse{
					StatusCode: tt.mockResponse.StatusCode,
					Body:       string(body),
				},
			)

			client, err := graphql.NewClient("https://api.upwork.com/graphql", recorder)
			require.NoError(t, err)

			service := NewBaseService(client, mocks.NewMockRateLimiter())

			var result map[string]interface{}
			err = service.ExecuteQuery(context.Background(), `{ test }`, nil, &result)
			
			assert.Error(t, err)
			if tt.checkType != nil {
				assert.True(t, tt.checkType(err), "Error should be of expected type")
			}
		})
	}
}