package alert

import (
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	capacity   int
	refillRate int
	lastRefill time.Time
	interval   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: capacity,
		lastRefill: time.Now(),
		interval:   interval,
	}
}

// Allow checks if an operation is allowed and consumes a token if it is
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// refill adds tokens based on time elapsed
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	if elapsed >= r.interval {
		// Full refill after interval
		r.tokens = r.capacity
		r.lastRefill = now
	} else {
		// Partial refill based on elapsed time
		tokensToAdd := int(float64(r.refillRate) * (elapsed.Seconds() / r.interval.Seconds()))
		if tokensToAdd > 0 {
			r.tokens = min(r.tokens+tokensToAdd, r.capacity)
			r.lastRefill = now
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

