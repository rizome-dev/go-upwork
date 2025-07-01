package mocks

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/stretchr/testify/mock"
)

// HTTPClient is a mock implementation of http.Client
type HTTPClient struct {
	mock.Mock
}

// Do mocks the Do method of http.Client
func (m *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if resp := args.Get(0); resp != nil {
		return resp.(*http.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

// HTTPRoundTripper is a mock implementation of http.RoundTripper
type HTTPRoundTripper struct {
	mock.Mock
}

// RoundTrip mocks the RoundTrip method
func (m *HTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if resp := args.Get(0); resp != nil {
		return resp.(*http.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockHTTPClient provides a simple mock HTTP client for testing
type MockHTTPClient struct {
	Responses []MockResponse
	CallCount int
}

// MockResponse represents a mocked HTTP response
type MockResponse struct {
	StatusCode int
	Body       string
	Headers    http.Header
	Error      error
}

// Do implements the http.Client Do method
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.CallCount >= len(m.Responses) {
		return nil, http.ErrNotSupported
	}
	
	resp := m.Responses[m.CallCount]
	m.CallCount++
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	headers := resp.Headers
	if headers == nil {
		headers = make(http.Header)
	}
	
	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(strings.NewReader(resp.Body)),
		Header:     headers,
		Request:    req,
	}, nil
}

// RequestRecorder records HTTP requests for testing
type RequestRecorder struct {
	Requests  []*http.Request
	Responses []MockResponse
	CallCount int
}

// NewRequestRecorder creates a new request recorder
func NewRequestRecorder(responses ...MockResponse) *RequestRecorder {
	return &RequestRecorder{
		Requests:  make([]*http.Request, 0),
		Responses: responses,
	}
}

// Do records the request and returns a mocked response
func (r *RequestRecorder) Do(req *http.Request) (*http.Response, error) {
	// Clone the request body so it can be read later
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	
	// Clone the request
	clonedReq := req.Clone(req.Context())
	if bodyBytes != nil {
		clonedReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	r.Requests = append(r.Requests, clonedReq)
	
	if r.CallCount >= len(r.Responses) {
		return nil, http.ErrNotSupported
	}
	
	resp := r.Responses[r.CallCount]
	r.CallCount++
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	headers := resp.Headers
	if headers == nil {
		headers = make(http.Header)
		headers.Set("Content-Type", "application/json")
	}
	
	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(strings.NewReader(resp.Body)),
		Header:     headers,
		Request:    req,
	}, nil
}

// GetLastRequest returns the last recorded request
func (r *RequestRecorder) GetLastRequest() *http.Request {
	if len(r.Requests) == 0 {
		return nil
	}
	return r.Requests[len(r.Requests)-1]
}

// GetRequestBody returns the body of a recorded request as a string
func (r *RequestRecorder) GetRequestBody(index int) string {
	if index >= len(r.Requests) {
		return ""
	}
	
	req := r.Requests[index]
	if req.Body == nil {
		return ""
	}
	
	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes)
}