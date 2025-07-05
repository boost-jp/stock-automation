package client

import (
	"context"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name string
		rps  int
	}{
		{
			name: "Create rate limiter with 10 RPS",
			rps:  10,
		},
		{
			name: "Create rate limiter with 1 RPS",
			rps:  1,
		},
		{
			name: "Create rate limiter with 100 RPS",
			rps:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rps)
			if rl == nil {
				t.Fatal("Expected non-nil rate limiter")
			}
			if rl.limiter == nil {
				t.Fatal("Expected non-nil limiter")
			}
		})
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	tests := []struct {
		name        string
		rps         int
		requests    int
		expectDelay bool
	}{
		{
			name:        "Single request should not delay",
			rps:         10,
			requests:    1,
			expectDelay: false,
		},
		{
			name:        "Multiple requests within rate limit",
			rps:         10,
			requests:    5,
			expectDelay: false,
		},
		{
			name:        "Requests exceeding rate limit should delay",
			rps:         2,
			requests:    3,
			expectDelay: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rps)
			ctx := context.Background()

			start := time.Now()
			for i := 0; i < tt.requests; i++ {
				if err := rl.Wait(ctx); err != nil {
					t.Fatalf("Wait() error = %v", err)
				}
			}
			elapsed := time.Since(start)

			// If we expect delay, elapsed time should be > 0
			if tt.expectDelay && elapsed < 100*time.Millisecond {
				t.Errorf("Expected delay but requests completed too quickly: %v", elapsed)
			}
		})
	}
}

func TestRateLimiter_Wait_WithCancelledContext(t *testing.T) {
	rl := NewRateLimiter(1)
	ctx, cancel := context.WithCancel(context.Background())

	// Make first request to consume the token
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("First Wait() error = %v", err)
	}

	// Cancel context
	cancel()

	// Next request should fail due to cancelled context
	err := rl.Wait(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}

func TestRateLimiter_TryWait(t *testing.T) {
	tests := []struct {
		name          string
		rps           int
		attempts      int
		expectSuccess []bool
	}{
		{
			name:          "First attempt should succeed",
			rps:           1,
			attempts:      1,
			expectSuccess: []bool{true},
		},
		{
			name:          "Multiple attempts with low rate limit",
			rps:           1,
			attempts:      3,
			expectSuccess: []bool{true, false, false}, // Only first succeeds immediately
		},
		{
			name:          "Multiple attempts with high rate limit",
			rps:           100,
			attempts:      3,
			expectSuccess: []bool{true, true, true}, // All should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rps)

			for i := 0; i < tt.attempts; i++ {
				got := rl.TryWait()
				if len(tt.expectSuccess) > i && got != tt.expectSuccess[i] {
					t.Errorf("Attempt %d: TryWait() = %v, want %v", i+1, got, tt.expectSuccess[i])
				}
			}
		})
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(10)
	ctx := context.Background()

	// Launch multiple goroutines
	done := make(chan bool, 20)
	start := time.Now()

	for i := 0; i < 20; i++ {
		go func() {
			err := rl.Wait(ctx)
			if err != nil {
				t.Errorf("Concurrent Wait() error = %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	elapsed := time.Since(start)
	// With 10 RPS and 20 requests, it should take at least 1 second
	if elapsed < 900*time.Millisecond {
		t.Errorf("Concurrent requests completed too quickly: %v", elapsed)
	}
}
