package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SlackAlertService implements the alert service using Slack
type SlackAlertService struct {
	webhookURL    string
	client        *http.Client
	rateLimiter   *RateLimiter
	mu            sync.Mutex
	alertHistory  map[string]time.Time
	alertInterval time.Duration
}

// SlackMessage represents a Slack message with attachments
type SlackMessage struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string       `json:"color"`
	Title      string       `json:"title"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields"`
	Footer     string       `json:"footer,omitempty"`
	FooterIcon string       `json:"footer_icon,omitempty"`
	Timestamp  int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackAlertService creates a new Slack alert service
func NewSlackAlertService() *SlackAlertService {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		logrus.Warn("SLACK_WEBHOOK_URL not set for alert service")
	}

	return &SlackAlertService{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter:   NewRateLimiter(10, time.Minute), // 10 alerts per minute
		alertHistory:  make(map[string]time.Time),
		alertInterval: 5 * time.Minute, // Same alert can be sent again after 5 minutes
	}
}

// Send sends an alert to Slack
func (s *SlackAlertService) Send(ctx context.Context, alert *Alert) error {
	if s.webhookURL == "" {
		logrus.WithFields(logrus.Fields{
			"level":   alert.Level.String(),
			"title":   alert.Title,
			"message": alert.Message,
			"error":   alert.Error,
		}).Error("Alert not sent: Slack webhook URL not configured")
		return nil
	}

	// Check if we should rate limit this alert
	if !s.shouldSendAlert(alert) {
		logrus.WithField("title", alert.Title).Debug("Alert rate limited")
		return nil
	}

	// Check rate limiter
	if !s.rateLimiter.Allow() {
		logrus.Warn("Alert rate limit exceeded")
		return fmt.Errorf("rate limit exceeded")
	}

	// Create Slack message
	msg := s.createSlackMessage(alert)

	// Send to Slack
	if err := s.sendSlackMessage(msg); err != nil {
		logrus.WithError(err).Error("Failed to send alert to Slack")
		return err
	}

	// Record that we sent this alert
	s.recordAlertSent(alert)

	// Log the alert
	s.logAlert(alert)

	return nil
}

// SendCritical sends a critical alert
func (s *SlackAlertService) SendCritical(ctx context.Context, title, message string, err error) error {
	alert := NewAlert(LevelCritical, title, message, err)
	return s.Send(ctx, alert)
}

// SendError sends an error alert
func (s *SlackAlertService) SendError(ctx context.Context, title, message string, err error) error {
	alert := NewAlert(LevelError, title, message, err)
	return s.Send(ctx, alert)
}

// SendWarning sends a warning alert
func (s *SlackAlertService) SendWarning(ctx context.Context, title, message string, err error) error {
	alert := NewAlert(LevelWarning, title, message, err)
	return s.Send(ctx, alert)
}

// shouldSendAlert checks if we should send this alert based on frequency limits
func (s *SlackAlertService) shouldSendAlert(alert *Alert) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", alert.Level.String(), alert.Title)
	lastSent, exists := s.alertHistory[key]

	if !exists {
		return true
	}

	return time.Since(lastSent) >= s.alertInterval
}

// recordAlertSent records that an alert was sent
func (s *SlackAlertService) recordAlertSent(alert *Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", alert.Level.String(), alert.Title)
	s.alertHistory[key] = time.Now()
}

// createSlackMessage creates a Slack message from an alert
func (s *SlackAlertService) createSlackMessage(alert *Alert) SlackMessage {
	color := s.getAlertColor(alert.Level)
	emoji := s.getAlertEmoji(alert.Level)

	fields := []SlackField{
		{
			Title: "Level",
			Value: alert.Level.String(),
			Short: true,
		},
		{
			Title: "Time",
			Value: alert.Timestamp.Format("2006-01-02 15:04:05"),
			Short: true,
		},
	}

	// Add error details if present
	if alert.Error != nil {
		fields = append(fields, SlackField{
			Title: "Error",
			Value: fmt.Sprintf("```%v```", alert.Error),
			Short: false,
		})
	}

	// Add context fields
	for key, value := range alert.Context {
		fields = append(fields, SlackField{
			Title: key,
			Value: fmt.Sprintf("%v", value),
			Short: true,
		})
	}

	attachment := SlackAttachment{
		Color:      color,
		Title:      fmt.Sprintf("%s %s", emoji, alert.Title),
		Text:       alert.Message,
		Fields:     fields,
		Footer:     "Stock Automation Alert System",
		FooterIcon: "https://platform.slack-edge.com/img/default_application_icon.png",
		Timestamp:  alert.Timestamp.Unix(),
	}

	return SlackMessage{
		Text:        fmt.Sprintf("%s Alert: %s", emoji, alert.Title),
		Attachments: []SlackAttachment{attachment},
	}
}

// getAlertColor returns the color for the alert level
func (s *SlackAlertService) getAlertColor(level Level) string {
	switch level {
	case LevelCritical:
		return "danger"
	case LevelError:
		return "warning"
	case LevelWarning:
		return "#ffcc00"
	default:
		return "good"
	}
}

// getAlertEmoji returns the emoji for the alert level
func (s *SlackAlertService) getAlertEmoji(level Level) string {
	switch level {
	case LevelCritical:
		return "🚨"
	case LevelError:
		return "❌"
	case LevelWarning:
		return "⚠️"
	default:
		return "ℹ️"
	}
}

// sendSlackMessage sends a message to Slack
func (s *SlackAlertService) sendSlackMessage(msg SlackMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Stock-Automation-Alert/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// logAlert logs the alert details
func (s *SlackAlertService) logAlert(alert *Alert) {
	fields := logrus.Fields{
		"level":     alert.Level.String(),
		"title":     alert.Title,
		"message":   alert.Message,
		"timestamp": alert.Timestamp,
	}

	if alert.Error != nil {
		fields["error"] = alert.Error.Error()
	}

	for key, value := range alert.Context {
		fields[fmt.Sprintf("context_%s", key)] = value
	}

	entry := logrus.WithFields(fields)

	switch alert.Level {
	case LevelCritical:
		entry.Error("Critical alert")
	case LevelError:
		entry.Error("Error alert")
	case LevelWarning:
		entry.Warn("Warning alert")
	default:
		entry.Info("Alert")
	}
}
