package alert

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// MockAlertService is a mock implementation of the alert service for testing
type MockAlertService struct {
	sentAlerts []*Alert
}

func (m *MockAlertService) Send(ctx context.Context, alert *Alert) error {
	m.sentAlerts = append(m.sentAlerts, alert)
	return nil
}

func (m *MockAlertService) SendCritical(ctx context.Context, title, message string, err error) error {
	return m.Send(ctx, NewAlert(LevelCritical, title, message, err))
}

func (m *MockAlertService) SendError(ctx context.Context, title, message string, err error) error {
	return m.Send(ctx, NewAlert(LevelError, title, message, err))
}

func (m *MockAlertService) SendWarning(ctx context.Context, title, message string, err error) error {
	return m.Send(ctx, NewAlert(LevelWarning, title, message, err))
}

func TestRecoveryMiddleware_Recover(t *testing.T) {
	mockService := &MockAlertService{}
	middleware := NewRecoveryMiddleware(mockService)

	// Function that will panic
	panicFunc := func() {
		panic("test panic")
	}

	// Wrap with recovery
	func() {
		defer middleware.Recover(context.Background(), "test operation")
		panicFunc()
	}()

	// Verify alert was sent
	if len(mockService.sentAlerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(mockService.sentAlerts))
	}

	alert := mockService.sentAlerts[0]
	if alert.Level != LevelCritical {
		t.Errorf("Expected critical level, got %v", alert.Level)
	}

	if alert.Title != "System Panic Detected" {
		t.Errorf("Expected 'System Panic Detected' title, got %v", alert.Title)
	}

	if !strings.Contains(alert.Message, "test operation") {
		t.Errorf("Alert message should contain operation name")
	}

	// Check context
	if alert.Context["operation"] != "test operation" {
		t.Errorf("Expected operation in context")
	}

	if _, ok := alert.Context["stack_trace"]; !ok {
		t.Error("Expected stack trace in context")
	}

	if alert.Context["panic_value"] != "test panic" {
		t.Errorf("Expected panic value in context")
	}
}

func TestRecoveryMiddleware_WrapOperation(t *testing.T) {
	mockService := &MockAlertService{}
	middleware := NewRecoveryMiddleware(mockService)

	tests := []struct {
		name        string
		operation   string
		fn          func() error
		expectAlert bool
		expectLevel Level
		wantErr     bool
	}{
		{
			name:      "Successful operation",
			operation: "success op",
			fn: func() error {
				return nil
			},
			expectAlert: false,
			wantErr:     false,
		},
		{
			name:      "Operation with non-critical error",
			operation: "error op",
			fn: func() error {
				return errors.New("simple error")
			},
			expectAlert: false,
			wantErr:     true,
		},
		{
			name:      "Operation with critical error",
			operation: "critical op",
			fn: func() error {
				return errors.New("database connection refused")
			},
			expectAlert: true,
			expectLevel: LevelCritical,
			wantErr:     true,
		},
		{
			name:      "Operation that panics",
			operation: "panic op",
			fn: func() error {
				panic("unexpected panic")
			},
			expectAlert: true,
			expectLevel: LevelCritical,
			wantErr:     false, // Panic is recovered, no error returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset alerts
			mockService.sentAlerts = nil

			// Execute operation
			err := middleware.WrapOperation(context.Background(), tt.operation, tt.fn)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("WrapOperation() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check alerts
			if tt.expectAlert {
				if len(mockService.sentAlerts) == 0 {
					t.Error("Expected alert to be sent")
				} else {
					alert := mockService.sentAlerts[0]
					if alert.Level != tt.expectLevel {
						t.Errorf("Expected alert level %v, got %v", tt.expectLevel, alert.Level)
					}
				}
			} else {
				if len(mockService.sentAlerts) > 0 {
					t.Error("No alert should be sent")
				}
			}
		})
	}
}

func TestIsCriticalError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		critical bool
	}{
		{
			name:     "Database error",
			err:      errors.New("database connection failed"),
			critical: true,
		},
		{
			name:     "Connection refused",
			err:      errors.New("dial tcp: connection refused"),
			critical: true,
		},
		{
			name:     "Timeout error",
			err:      errors.New("operation timeout"),
			critical: true,
		},
		{
			name:     "Unauthorized error",
			err:      errors.New("unauthorized access"),
			critical: true,
		},
		{
			name:     "Fatal error",
			err:      errors.New("fatal: cannot proceed"),
			critical: true,
		},
		{
			name:     "Simple error",
			err:      errors.New("file not found"),
			critical: false,
		},
		{
			name:     "Validation error",
			err:      errors.New("invalid input"),
			critical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCriticalError(tt.err); got != tt.critical {
				t.Errorf("isCriticalError() = %v, want %v", got, tt.critical)
			}
		})
	}
}
