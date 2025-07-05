package alert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSlackAlertService_Send(t *testing.T) {
	// Create a test server to mock Slack webhook
	var receivedPayload SlackMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if ct := r.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got %s", ct)
		}

		// Decode the payload
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set webhook URL to test server
	os.Setenv("SLACK_WEBHOOK_URL", server.URL)
	defer os.Unsetenv("SLACK_WEBHOOK_URL")

	// Create service
	service := NewSlackAlertService()
	service.webhookURL = server.URL

	// Test sending different alert levels
	tests := []struct {
		name      string
		alert     *Alert
		wantColor string
		wantEmoji string
	}{
		{
			name:      "Critical alert",
			alert:     NewAlert(LevelCritical, "Critical Test", "This is critical", errors.New("critical error")),
			wantColor: "danger",
			wantEmoji: "🚨",
		},
		{
			name:      "Error alert",
			alert:     NewAlert(LevelError, "Error Test", "This is an error", errors.New("error")),
			wantColor: "warning",
			wantEmoji: "❌",
		},
		{
			name:      "Warning alert",
			alert:     NewAlert(LevelWarning, "Warning Test", "This is a warning", nil),
			wantColor: "#ffcc00",
			wantEmoji: "⚠️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add context to alert
			tt.alert.WithContext("test_key", "test_value")

			// Send alert
			if err := service.Send(context.Background(), tt.alert); err != nil {
				t.Errorf("Failed to send alert: %v", err)
			}

			// Verify payload
			if len(receivedPayload.Attachments) != 1 {
				t.Fatalf("Expected 1 attachment, got %d", len(receivedPayload.Attachments))
			}

			attachment := receivedPayload.Attachments[0]
			if attachment.Color != tt.wantColor {
				t.Errorf("Expected color %s, got %s", tt.wantColor, attachment.Color)
			}

			// Check that emoji is in the title
			if !containsIgnoreCase(attachment.Title, tt.wantEmoji) {
				t.Errorf("Expected emoji %s in title, got %s", tt.wantEmoji, attachment.Title)
			}
		})
	}
}

func TestSlackAlertService_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSlackAlertService()
	service.webhookURL = server.URL
	service.rateLimiter = NewRateLimiter(2, 100*time.Millisecond)
	service.alertInterval = 0 // Disable frequency limiting for this test

	ctx := context.Background()

	// First two alerts should succeed
	for i := 0; i < 2; i++ {
		if err := service.SendError(ctx, fmt.Sprintf("Test %d", i), "Message", nil); err != nil {
			t.Errorf("Alert %d should succeed: %v", i+1, err)
		}
	}

	// Third alert should be rate limited
	if err := service.SendError(ctx, "Test 3", "Message", nil); err == nil {
		t.Error("Third alert should be rate limited")
	}
}

func TestSlackAlertService_FrequencyLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSlackAlertService()
	service.webhookURL = server.URL
	service.alertInterval = 100 * time.Millisecond

	ctx := context.Background()
	alert := NewAlert(LevelError, "Same Alert", "Same message", nil)

	// First alert should succeed
	if err := service.Send(ctx, alert); err != nil {
		t.Errorf("First alert should succeed: %v", err)
	}

	// Immediate second alert should be skipped
	if err := service.Send(ctx, alert); err != nil {
		t.Errorf("Second alert should be skipped without error: %v", err)
	}

	// Wait for interval to pass
	time.Sleep(110 * time.Millisecond)

	// Alert should be allowed again
	if err := service.Send(ctx, alert); err != nil {
		t.Errorf("Alert after interval should succeed: %v", err)
	}
}

func TestSlackAlertService_NoWebhookURL(t *testing.T) {
	service := NewSlackAlertService()
	service.webhookURL = ""

	// Should not return error when webhook URL is not set
	if err := service.SendCritical(context.Background(), "Test", "Message", nil); err != nil {
		t.Errorf("Should not return error when webhook URL is not set: %v", err)
	}
}
