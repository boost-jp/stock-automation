package client

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"syscall"
)

// Error types for better error handling and retry logic
var (
	// Retryable errors
	ErrRateLimit    = errors.New("rate limit exceeded")
	ErrServerError  = errors.New("server error")
	ErrTimeout      = errors.New("request timeout")
	ErrNetworkError = errors.New("network error")

	// Non-retryable errors
	ErrNotFound     = errors.New("resource not found")
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
)

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	if errors.Is(err, ErrRateLimit) || errors.Is(err, ErrServerError) ||
		errors.Is(err, ErrTimeout) || errors.Is(err, ErrNetworkError) {
		return true
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	// Check for URL errors (DNS failures, etc.)
	if urlErr, ok := err.(*url.Error); ok {
		return IsRetryableError(urlErr.Err)
	}

	// Check for syscall errors
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	// Check error messages for common retryable scenarios
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"temporary failure",
		"connection reset",
		"connection refused",
		"no such host",
		"too many requests",
		"service unavailable",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// ClassifyHTTPError converts HTTP status codes to appropriate error types
func ClassifyHTTPError(statusCode int) error {
	switch statusCode {
	case 404:
		return ErrNotFound
	case 400:
		return ErrBadRequest
	case 401, 403:
		return ErrUnauthorized
	case 429:
		return ErrRateLimit
	case 500, 502, 503, 504:
		return ErrServerError
	default:
		if statusCode >= 500 {
			return ErrServerError
		}
		return nil
	}
}
