package alert

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter with capacity 3, refilling every 100ms
	rl := NewRateLimiter(3, 100*time.Millisecond)

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow() {
		t.Error("4th request should be denied")
	}

	// Wait for refill
	time.Sleep(110 * time.Millisecond)

	// Should allow requests again
	if !rl.Allow() {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimiter_PartialRefill(t *testing.T) {
	// Create a rate limiter with capacity 10, refilling every second
	rl := NewRateLimiter(10, time.Second)

	// Use all tokens
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Errorf("Initial request %d should be allowed", i+1)
		}
	}

	// Should be denied
	if rl.Allow() {
		t.Error("Request should be denied when tokens exhausted")
	}

	// Wait for partial refill (half the interval)
	time.Sleep(500 * time.Millisecond)

	// Should have about 5 tokens refilled
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}

	// Should allow approximately 5 requests (±1 for timing variations)
	if allowed < 4 || allowed > 6 {
		t.Errorf("Expected around 5 requests to be allowed after partial refill, got %d", allowed)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, time.Second)

	// Run concurrent goroutines trying to consume tokens
	done := make(chan bool, 10)
	allowed := make(chan bool, 1000)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				if rl.Allow() {
					allowed <- true
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	close(allowed)

	// Count allowed requests
	count := 0
	for range allowed {
		count++
	}

	// Should have allowed exactly 100 requests
	if count != 100 {
		t.Errorf("Expected 100 requests to be allowed, got %d", count)
	}
}

