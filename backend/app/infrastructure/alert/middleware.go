package alert

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// RecoveryMiddleware provides panic recovery and error alerting
type RecoveryMiddleware struct {
	alertService Service
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(alertService Service) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		alertService: alertService,
	}
}

// Recover recovers from panics and sends alerts
func (m *RecoveryMiddleware) Recover(ctx context.Context, operation string) {
	if r := recover(); r != nil {
		// Get stack trace
		stack := string(debug.Stack())

		// Create panic error
		panicErr := fmt.Errorf("panic in %s: %v", operation, r)

		// Send critical alert
		alert := NewAlert(LevelCritical, "System Panic Detected", fmt.Sprintf("A panic occurred in operation: %s", operation), panicErr)
		alert.WithContext("stack_trace", stack)
		alert.WithContext("operation", operation)
		alert.WithContext("panic_value", fmt.Sprintf("%v", r))

		// Try to send alert (don't panic if this fails)
		if err := m.alertService.Send(ctx, alert); err != nil {
			logrus.WithError(err).Error("Failed to send panic alert")
		}

		// Log the panic
		logrus.WithFields(logrus.Fields{
			"operation":   operation,
			"panic":       r,
			"stack_trace": stack,
		}).Error("Panic recovered")
	}
}

// WrapOperation wraps an operation with panic recovery
func (m *RecoveryMiddleware) WrapOperation(ctx context.Context, operation string, fn func() error) error {
	defer m.Recover(ctx, operation)

	// Execute the operation
	if err := fn(); err != nil {
		// Check if this is a critical error
		if isCriticalError(err) {
			alert := NewAlert(LevelCritical, "Critical Error", fmt.Sprintf("Critical error in operation: %s", operation), err)
			alert.WithContext("operation", operation)

			// Send alert (don't fail the operation if alert fails)
			if alertErr := m.alertService.Send(ctx, alert); alertErr != nil {
				logrus.WithError(alertErr).Error("Failed to send critical error alert")
			}
		}

		return err
	}

	return nil
}

// isCriticalError determines if an error is critical
func isCriticalError(err error) bool {
	// You can customize this logic based on your error types
	// For now, we'll consider certain error strings as critical
	errorStr := err.Error()

	criticalPatterns := []string{
		"database",
		"connection refused",
		"timeout",
		"unauthorized",
		"fatal",
		"panic",
	}

	for _, pattern := range criticalPatterns {
		if containsIgnoreCase(errorStr, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsIgnoreCaseHelper(s, substr)
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + 'a' - 'A'
	}
	return b
}
