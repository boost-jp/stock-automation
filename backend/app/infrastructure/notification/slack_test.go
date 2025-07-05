package notification

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/stretchr/testify/assert"
)

func TestSlackNotifier_SendMessage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	err := notifier.SendMessage("Test message")
	assert.NoError(t, err)
}

func TestSlackNotifier_SendMessage_NoWebhookURL(t *testing.T) {
	notifier := &SlackNotifier{
		webhookURL: "",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	err := notifier.SendMessage("Test message")
	assert.NoError(t, err) // Should not error when webhook URL is empty
}

func TestSlackNotifier_SendStockAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	tests := []struct {
		name       string
		alertType  string
		stockCode  string
		stockName  string
		currentPrice float64
		targetPrice  float64
	}{
		{"Buy Alert", "buy", "7203", "Toyota", 2500.0, 2400.0},
		{"Sell Alert", "sell", "9983", "Fast Retailing", 85000.0, 86000.0},
		{"Warning Alert", "warning", "6758", "Sony", 15000.0, 15500.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := notifier.SendStockAlert(tt.stockCode, tt.stockName, tt.currentPrice, tt.targetPrice, tt.alertType)
			assert.NoError(t, err)
		})
	}
}

func TestSlackNotifier_SendDailyReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	tests := []struct {
		name        string
		totalValue  float64
		totalGain   float64
		gainPercent float64
	}{
		{"Positive Gain", 1000000.0, 50000.0, 5.0},
		{"Negative Gain", 900000.0, -100000.0, -10.0},
		{"Zero Gain", 1000000.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := notifier.SendDailyReport(tt.totalValue, tt.totalGain, tt.gainPercent)
			assert.NoError(t, err)
		})
	}
}

func TestSlackNotifier_SendComprehensiveReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	summary := &domain.PortfolioSummary{
		TotalValue:       1500000.0,
		TotalCost:        1400000.0,
		TotalGain:        100000.0,
		TotalGainPercent: 7.14,
		Holdings: []domain.HoldingSummary{
			{
				Code:         "7203",
				Name:         "Toyota",
				Shares:       100,
				CurrentPrice: 2500.0,
				CurrentValue: 250000.0,
				Gain:         10000.0,
				GainPercent:  4.17,
			},
			{
				Code:         "9983",
				Name:         "Fast Retailing",
				Shares:       10,
				CurrentPrice: 85000.0,
				CurrentValue: 850000.0,
				Gain:         -10000.0,
				GainPercent:  -1.16,
			},
		},
	}

	report := "Test comprehensive report"
	err := notifier.SendComprehensiveReport(report, summary)
	assert.NoError(t, err)
}

func TestSlackNotifier_RetryMechanism(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	err := notifier.SendMessage("Test retry")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestSlackNotifier_RetryFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 2,
		retryDelay: 100 * time.Millisecond,
	}

	err := notifier.SendMessage("Test fail")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Slack notification after 3 attempts")
}

func TestSlackNotifier_NetworkError(t *testing.T) {
	notifier := &SlackNotifier{
		webhookURL: "http://invalid-url-that-does-not-exist.com/webhook",
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		maxRetries: 1,
		retryDelay: 100 * time.Millisecond,
	}

	err := notifier.SendMessage("Test network error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Slack notification")
}

// MockNotificationLogRepository for testing
type MockNotificationLogRepository struct {
	CreateFunc      func(ctx context.Context, log *repository.NotificationLog) error
	UpdateStatusFunc func(ctx context.Context, id int64, status string, errorMessage *string, sentAt *time.Time) error
}

func (m *MockNotificationLogRepository) Create(ctx context.Context, log *repository.NotificationLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	log.ID = 1
	return nil
}

func (m *MockNotificationLogRepository) UpdateStatus(ctx context.Context, id int64, status string, errorMessage *string, sentAt *time.Time) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, errorMessage, sentAt)
	}
	return nil
}

func (m *MockNotificationLogRepository) GetRecent(ctx context.Context, limit int) ([]*repository.NotificationLog, error) {
	return nil, nil
}

func (m *MockNotificationLogRepository) GetByType(ctx context.Context, notificationType string, limit int) ([]*repository.NotificationLog, error) {
	return nil, nil
}

func TestSlackNotifier_WithLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	var createdLog *repository.NotificationLog
	var updatedID int64
	var updatedStatus string

	mockRepo := &MockNotificationLogRepository{
		CreateFunc: func(ctx context.Context, log *repository.NotificationLog) error {
			createdLog = log
			log.ID = 123
			return nil
		},
		UpdateStatusFunc: func(ctx context.Context, id int64, status string, errorMessage *string, sentAt *time.Time) error {
			updatedID = id
			updatedStatus = status
			return nil
		},
	}

	notifier := &SlackNotifier{
		webhookURL: server.URL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
		logRepo:    mockRepo,
	}

	err := notifier.SendMessage("Test with logging")
	assert.NoError(t, err)
	assert.NotNil(t, createdLog)
	assert.Equal(t, "message", createdLog.NotificationType)
	assert.Equal(t, "pending", createdLog.Status)
	assert.Equal(t, int64(123), updatedID)
	assert.Equal(t, "sent", updatedStatus)
}