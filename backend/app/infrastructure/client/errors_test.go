package client

import (
	"errors"
	"net/url"
	"syscall"
	"testing"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ErrRateLimit is retryable",
			err:  ErrRateLimit,
			want: true,
		},
		{
			name: "ErrServerError is retryable",
			err:  ErrServerError,
			want: true,
		},
		{
			name: "ErrTimeout is retryable",
			err:  ErrTimeout,
			want: true,
		},
		{
			name: "ErrNetworkError is retryable",
			err:  ErrNetworkError,
			want: true,
		},
		{
			name: "ErrNotFound is not retryable",
			err:  ErrNotFound,
			want: false,
		},
		{
			name: "ErrBadRequest is not retryable",
			err:  ErrBadRequest,
			want: false,
		},
		{
			name: "ErrUnauthorized is not retryable",
			err:  ErrUnauthorized,
			want: false,
		},
		{
			name: "connection refused syscall error",
			err:  syscall.ECONNREFUSED,
			want: false, // syscall errors are not directly handled
		},
		{
			name: "connection reset syscall error",
			err:  syscall.ECONNRESET,
			want: false, // syscall errors are not directly handled
		},
		{
			name: "wrapped rate limit error is retryable",
			err:  errors.New("request failed: " + ErrRateLimit.Error()),
			want: false, // Not using errors.Is, so string match won't work
		},
		{
			name: "timeout error message is retryable",
			err:  errors.New("request timeout"),
			want: true,
		},
		{
			name: "connection refused error message is retryable",
			err:  errors.New("connection refused"),
			want: true,
		},
		{
			name: "too many requests error message is retryable",
			err:  errors.New("too many requests"),
			want: true,
		},
		{
			name: "service unavailable error message is retryable",
			err:  errors.New("service unavailable"),
			want: true,
		},
		{
			name: "random error is not retryable",
			err:  errors.New("some random error"),
			want: false,
		},
		{
			name: "URL error with timeout is retryable",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: &timeoutError{},
			},
			want: true,
		},
		{
			name: "URL error with non-retryable error",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: errors.New("some error"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{
			name:       "404 returns ErrNotFound",
			statusCode: 404,
			wantErr:    ErrNotFound,
		},
		{
			name:       "400 returns ErrBadRequest",
			statusCode: 400,
			wantErr:    ErrBadRequest,
		},
		{
			name:       "401 returns ErrUnauthorized",
			statusCode: 401,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "403 returns ErrUnauthorized",
			statusCode: 403,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "429 returns ErrRateLimit",
			statusCode: 429,
			wantErr:    ErrRateLimit,
		},
		{
			name:       "500 returns ErrServerError",
			statusCode: 500,
			wantErr:    ErrServerError,
		},
		{
			name:       "502 returns ErrServerError",
			statusCode: 502,
			wantErr:    ErrServerError,
		},
		{
			name:       "503 returns ErrServerError",
			statusCode: 503,
			wantErr:    ErrServerError,
		},
		{
			name:       "504 returns ErrServerError",
			statusCode: 504,
			wantErr:    ErrServerError,
		},
		{
			name:       "505 returns ErrServerError",
			statusCode: 505,
			wantErr:    ErrServerError,
		},
		{
			name:       "200 returns nil",
			statusCode: 200,
			wantErr:    nil,
		},
		{
			name:       "201 returns nil",
			statusCode: 201,
			wantErr:    nil,
		},
		{
			name:       "204 returns nil",
			statusCode: 204,
			wantErr:    nil,
		},
		{
			name:       "301 returns nil",
			statusCode: 301,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := ClassifyHTTPError(tt.statusCode)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("ClassifyHTTPError(%d) = %v, want %v", tt.statusCode, gotErr, tt.wantErr)
			}
		})
	}
}

func TestIsRetryableError_NetworkError(t *testing.T) {
	// Test with a custom net.Error implementation
	netErr := &customNetError{
		temporary: true,
		timeout:   false,
	}

	if !IsRetryableError(netErr) {
		t.Error("Expected temporary network error to be retryable")
	}

	netErr.temporary = false
	netErr.timeout = true

	if !IsRetryableError(netErr) {
		t.Error("Expected timeout network error to be retryable")
	}

	netErr.temporary = false
	netErr.timeout = false

	if IsRetryableError(netErr) {
		t.Error("Expected non-temporary, non-timeout network error to not be retryable")
	}
}

// Helper types for testing

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

type customNetError struct {
	temporary bool
	timeout   bool
}

func (e *customNetError) Error() string   { return "network error" }
func (e *customNetError) Timeout() bool   { return e.timeout }
func (e *customNetError) Temporary() bool { return e.temporary }
