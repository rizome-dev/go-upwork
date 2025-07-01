package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockHTTPResponse creates a mock HTTP response with the given status code and body
func MockHTTPResponse(statusCode int, body interface{}) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
		case []byte:
			bodyReader = bytes.NewBuffer(v)
		default:
			b, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(b)
		}
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bodyReader),
		Header:     make(http.Header),
	}
}

// MockGraphQLResponse creates a mock GraphQL response
func MockGraphQLResponse(data interface{}, errors []interface{}) map[string]interface{} {
	response := make(map[string]interface{})
	if data != nil {
		response["data"] = data
	}
	if errors != nil {
		response["errors"] = errors
	}
	return response
}

// MockOAuth2Token creates a mock OAuth2 token response
func MockOAuth2Token(accessToken string, expiresIn int) map[string]interface{} {
	return map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    expiresIn,
		"refresh_token": "mock_refresh_token",
		"scope":        "read write",
	}
}

// AssertJSONEqual asserts that two JSON strings are equal
func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedObj, actualObj interface{}
	err := json.Unmarshal([]byte(expected), &expectedObj)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(actual), &actualObj)
	assert.NoError(t, err)
	assert.Equal(t, expectedObj, actualObj)
}

// MockTime provides a consistent time for testing
func MockTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// CreateTestServer creates a test HTTP server with custom handlers
type TestServer struct {
	*http.ServeMux
	server *http.Server
}

func NewTestServer() *TestServer {
	mux := http.NewServeMux()
	return &TestServer{
		ServeMux: mux,
	}
}

func (ts *TestServer) Start(address string) {
	ts.server = &http.Server{
		Addr:    address,
		Handler: ts.ServeMux,
	}
	go ts.server.ListenAndServe()
	time.Sleep(100 * time.Millisecond) // Give server time to start
}

func (ts *TestServer) Stop() {
	if ts.server != nil {
		ts.server.Close()
	}
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []map[string]int       `json:"locations,omitempty"`
}

// CreateGraphQLError creates a GraphQL error for testing
func CreateGraphQLError(message string, code string) GraphQLError {
	return GraphQLError{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

// SampleData provides sample data for testing various API responses
var SampleData = struct {
	UserID        string
	ContractID    string
	JobPostingID  string
	ProposalID    string
	MilestoneID   string
	OrganizationID string
	TeamID        string
	RoomID        string
}{
	UserID:        "123456789",
	ContractID:    "987654321",
	JobPostingID:  "456789123",
	ProposalID:    "789123456",
	MilestoneID:   "321654987",
	OrganizationID: "org_123456",
	TeamID:        "team_789",
	RoomID:        "room_456",
}

// MockContractResponse creates a mock contract response
func MockContractResponse() map[string]interface{} {
	return map[string]interface{}{
		"id": SampleData.ContractID,
		"title": "Test Contract",
		"status": "ACTIVE",
		"hourlyRate": map[string]interface{}{
			"amount":   "50.00",
			"currency": "USD",
		},
		"startDate": "2024-01-01T00:00:00Z",
		"client": map[string]interface{}{
			"id":   "client_123",
			"name": "Test Client",
		},
		"freelancer": map[string]interface{}{
			"id":   "freelancer_456",
			"name": "Test Freelancer",
		},
	}
}

// MockJobPostingResponse creates a mock job posting response
func MockJobPostingResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":    SampleData.JobPostingID,
		"title": "Test Job Posting",
		"description": "This is a test job posting",
		"budget": map[string]interface{}{
			"amount":   "1000.00",
			"currency": "USD",
		},
		"duration": "1-3 months",
		"experienceLevel": "INTERMEDIATE",
		"skills": []map[string]interface{}{
			{"name": "Go"},
			{"name": "GraphQL"},
			{"name": "Testing"},
		},
		"client": map[string]interface{}{
			"id":   "client_123",
			"name": "Test Client",
		},
	}
}

// MockProposalResponse creates a mock proposal response
func MockProposalResponse() map[string]interface{} {
	return map[string]interface{}{
		"id": SampleData.ProposalID,
		"jobPosting": map[string]interface{}{
			"id":    SampleData.JobPostingID,
			"title": "Test Job",
		},
		"coverLetter": "I am interested in this job...",
		"bidAmount": map[string]interface{}{
			"amount":   "50.00",
			"currency": "USD",
		},
		"status": "SUBMITTED",
		"freelancer": map[string]interface{}{
			"id":   "freelancer_456",
			"name": "Test Freelancer",
		},
	}
}

// MockUserResponse creates a mock user response
func MockUserResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":    SampleData.UserID,
		"email": "test@example.com",
		"firstName": "Test",
		"lastName": "User",
		"profile": map[string]interface{}{
			"title": "Software Developer",
			"overview": "Experienced developer",
			"hourlyRate": map[string]interface{}{
				"amount":   "75.00",
				"currency": "USD",
			},
		},
	}
}