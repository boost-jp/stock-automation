package alert

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "Critical level",
			level: LevelCritical,
			want:  "CRITICAL",
		},
		{
			name:  "Error level",
			level: LevelError,
			want:  "ERROR",
		},
		{
			name:  "Warning level",
			level: LevelWarning,
			want:  "WARNING",
		},
		{
			name:  "Unknown level",
			level: Level(999),
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAlert(t *testing.T) {
	err := errors.New("test error")

	alert := NewAlert(LevelError, "Test Alert", "This is a test message", err)

	if alert.Level != LevelError {
		t.Errorf("Expected level %v, got %v", LevelError, alert.Level)
	}

	if alert.Title != "Test Alert" {
		t.Errorf("Expected title 'Test Alert', got %v", alert.Title)
	}

	if alert.Message != "This is a test message" {
		t.Errorf("Expected message 'This is a test message', got %v", alert.Message)
	}

	if alert.Error != err {
		t.Errorf("Expected error %v, got %v", err, alert.Error)
	}

	if time.Since(alert.Timestamp) > time.Second {
		t.Errorf("Timestamp should be recent")
	}

	if alert.Context == nil {
		t.Error("Context should be initialized")
	}
}

func TestAlert_WithContext(t *testing.T) {
	alert := NewAlert(LevelWarning, "Test", "Message", nil)

	alert.WithContext("key1", "value1").
		WithContext("key2", 123).
		WithContext("key3", true)

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	if diff := cmp.Diff(expected, alert.Context); diff != "" {
		t.Errorf("Context mismatch (-want +got):\n%s", diff)
	}
}

