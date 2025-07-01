package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// RateLimiter is a mock implementation of the RateLimiter interface
type RateLimiter struct {
	mock.Mock
}

// Wait mocks the Wait method
func (m *RateLimiter) Wait(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockRateLimiter provides a simple mock rate limiter for testing
type MockRateLimiter struct {
	mu           sync.Mutex
	WaitCalls    int
	WaitDuration time.Duration
	ShouldError  bool
	Error        error
}

// NewMockRateLimiter creates a new mock rate limiter
func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{}
}

// Wait simulates rate limiting
func (m *MockRateLimiter) Wait(ctx context.Context) error {
	m.mu.Lock()
	m.WaitCalls++
	m.mu.Unlock()
	
	if m.ShouldError {
		return m.Error
	}
	
	if m.WaitDuration > 0 {
		select {
		case <-time.After(m.WaitDuration):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return nil
}

// GetWaitCalls returns the number of times Wait was called
func (m *MockRateLimiter) GetWaitCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.WaitCalls
}

// Reset resets the mock state
func (m *MockRateLimiter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WaitCalls = 0
	m.WaitDuration = 0
	m.ShouldError = false
	m.Error = nil
}

// NoOpRateLimiter is a rate limiter that does nothing
type NoOpRateLimiter struct{}

// Wait does nothing and returns immediately
func (n *NoOpRateLimiter) Wait(ctx context.Context) error {
	return nil
}

// RecordingRateLimiter records all wait calls with timestamps
type RecordingRateLimiter struct {
	mu         sync.Mutex
	WaitCalls  []time.Time
	ReturnError error
}

// NewRecordingRateLimiter creates a new recording rate limiter
func NewRecordingRateLimiter() *RecordingRateLimiter {
	return &RecordingRateLimiter{
		WaitCalls: make([]time.Time, 0),
	}
}

// Wait records the call time and returns the configured error
func (r *RecordingRateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	r.WaitCalls = append(r.WaitCalls, time.Now())
	r.mu.Unlock()
	
	return r.ReturnError
}

// GetCallCount returns the number of times Wait was called
func (r *RecordingRateLimiter) GetCallCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.WaitCalls)
}

// GetCallTimes returns all recorded call times
func (r *RecordingRateLimiter) GetCallTimes() []time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]time.Time, len(r.WaitCalls))
	copy(result, r.WaitCalls)
	return result
}

// Reset clears all recorded calls
func (r *RecordingRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.WaitCalls = make([]time.Time, 0)
	r.ReturnError = nil
}