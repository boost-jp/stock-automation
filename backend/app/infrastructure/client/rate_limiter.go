package client

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality for API calls
type RateLimiter struct {
	limiter   *rate.Limiter
	mu        sync.Mutex
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter with specified requests per second
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		limiter:   rate.NewLimiter(rate.Limit(rps), rps),
		lastReset: time.Now(),
	}
}

// Wait blocks until the rate limiter allows another request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// TryWait attempts to reserve a request slot without blocking
func (rl *RateLimiter) TryWait() bool {
	return rl.limiter.Allow()
}
