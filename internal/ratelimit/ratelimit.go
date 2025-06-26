// Package ratelimit provides rate limiting functionality for the Upwork SDK.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter
type Limiter struct {
	tokens    int
	maxTokens int
	interval  time.Duration
	lastReset time.Time
	mu        sync.Mutex
}

// New creates a new rate limiter
func New(maxRequests int, interval time.Duration) *Limiter {
	return &Limiter{
		tokens:    maxRequests,
		maxTokens: maxRequests,
		interval:  interval,
		lastReset: time.Now(),
	}
}

// Wait blocks until a token is available
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		
		// Reset tokens if interval has passed
		if time.Since(l.lastReset) >= l.interval {
			l.tokens = l.maxTokens
			l.lastReset = time.Now()
		}
		
		// If tokens available, consume one and return
		if l.tokens > 0 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		
		// Calculate wait time until next reset
		waitTime := l.interval - time.Since(l.lastReset)
		l.mu.Unlock()
		
		// Wait with context
		timer := time.NewTimer(waitTime)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue loop to try again
		}
	}
}

// Available returns the number of available tokens
func (l *Limiter) Available() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Reset tokens if interval has passed
	if time.Since(l.lastReset) >= l.interval {
		l.tokens = l.maxTokens
		l.lastReset = time.Now()
	}
	
	return l.tokens
}