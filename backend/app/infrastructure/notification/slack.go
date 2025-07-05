package notification

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/boost-jp/stock-automation/app/domain"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/sirupsen/logrus"
)

type SlackNotifier struct {
	webhookURL string
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
	logRepo    repository.NotificationLogRepository
}

type SlackMessage struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color  string       `json:"color"`
	Title  string       `json:"title"`
	Fields []SlackField `json:"fields"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NewSlackNotifier() *SlackNotifier {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		logrus.Warn("SLACK_WEBHOOK_URL not set")
	}

	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

// NewSlackNotificationService creates a new Slack notification service with explicit configuration
func NewSlackNotificationService(webhookURL, channel, username string) NotificationService {
	if webhookURL == "" {
		logrus.Warn("Slack webhook URL not set")
	}

	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

func (s *SlackNotifier) SendMessage(message string) error {
	if s.webhookURL == "" {
		logrus.Debug("Slack webhook URL not configured, skipping notification")
		return nil
	}

	msg := SlackMessage{
		Text: message,
	}

	return s.sendSlackMessageWithLog(context.Background(), msg, "message", nil)
}

func (s *SlackNotifier) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	if s.webhookURL == "" {
		return nil
	}

	color := "warning"
	if alertType == "buy" {
		color = "good"
	} else if alertType == "sell" {
		color = "danger"
	}

	msg := SlackMessage{
		Text: fmt.Sprintf("🔔 株価アラート: %s (%s)", stockName, stockCode),
		Attachments: []SlackAttachment{
			{
				Color: color,
				Title: fmt.Sprintf("%s通知", alertType),
				Fields: []SlackField{
					{
						Title: "現在価格",
						Value: fmt.Sprintf("¥%.2f", currentPrice),
						Short: true,
					},
					{
						Title: "目標価格",
						Value: fmt.Sprintf("¥%.2f", targetPrice),
						Short: true,
					},
					{
						Title: "乖離率",
						Value: fmt.Sprintf("%.2f%%", (currentPrice-targetPrice)/targetPrice*100),
						Short: true,
					},
					{
						Title: "時刻",
						Value: time.Now().Format("2006-01-02 15:04:05"),
						Short: true,
					},
				},
			},
		},
	}

	metadata := map[string]interface{}{
		"stock_code":    stockCode,
		"stock_name":    stockName,
		"current_price": currentPrice,
		"target_price":  targetPrice,
		"alert_type":    alertType,
	}

	return s.sendSlackMessageWithLog(context.Background(), msg, "stock_alert", metadata)
}

func (s *SlackNotifier) SendDailyReport(totalValue, totalGain float64, gainPercent float64) error {
	if s.webhookURL == "" {
		return nil
	}

	color := "good"
	if totalGain < 0 {
		color = "danger"
	}

	msg := SlackMessage{
		Text: "📊 本日の投資状況レポート",
		Attachments: []SlackAttachment{
			{
				Color: color,
				Title: "ポートフォリオ状況",
				Fields: []SlackField{
					{
						Title: "総資産",
						Value: fmt.Sprintf("¥%.2f", totalValue),
						Short: true,
					},
					{
						Title: "損益",
						Value: fmt.Sprintf("¥%.2f", totalGain),
						Short: true,
					},
					{
						Title: "損益率",
						Value: fmt.Sprintf("%.2f%%", gainPercent),
						Short: true,
					},
					{
						Title: "更新時刻",
						Value: time.Now().Format("2006-01-02 15:04:05"),
						Short: true,
					},
				},
			},
		},
	}

	metadata := map[string]interface{}{
		"total_value":  totalValue,
		"total_gain":   totalGain,
		"gain_percent": gainPercent,
	}

	return s.sendSlackMessageWithLog(context.Background(), msg, "daily_report", metadata)
}

// SendComprehensiveReport sends a comprehensive daily report with enhanced formatting
func (s *SlackNotifier) SendComprehensiveReport(report string, summary *domain.PortfolioSummary) error {
	if s.webhookURL == "" {
		return nil
	}

	color := "good"
	if summary.TotalGain < 0 {
		color = "danger"
	}

	// Create blocks for better formatting
	attachments := []SlackAttachment{
		{
			Color: color,
			Title: "📊 ポートフォリオサマリー",
			Fields: []SlackField{
				{
					Title: "総資産",
					Value: fmt.Sprintf("¥%,.0f", summary.TotalValue),
					Short: true,
				},
				{
					Title: "総投資額",
					Value: fmt.Sprintf("¥%,.0f", summary.TotalCost),
					Short: true,
				},
				{
					Title: "損益",
					Value: fmt.Sprintf("¥%,.0f", summary.TotalGain),
					Short: true,
				},
				{
					Title: "損益率",
					Value: fmt.Sprintf("%.2f%%", summary.TotalGainPercent),
					Short: true,
				},
			},
		},
	}

	// Add holdings details if available
	if len(summary.Holdings) > 0 {
		holdings := SlackAttachment{
			Color:  "info",
			Title:  "📈 保有銘柄詳細",
			Fields: []SlackField{},
		}

		for _, holding := range summary.Holdings {
			holdingColor := "🟢"
			if holding.Gain < 0 {
				holdingColor = "🔴"
			}

			holdings.Fields = append(holdings.Fields, SlackField{
				Title: fmt.Sprintf("%s %s (%s)", holdingColor, holding.Name, holding.Code),
				Value: fmt.Sprintf("数量: %d | 現在値: ¥%,.0f | 損益: ¥%,.0f (%.1f%%)",
					holding.Shares, holding.CurrentPrice, holding.Gain, holding.GainPercent),
				Short: false,
			})
		}

		attachments = append(attachments, holdings)
	}

	msg := SlackMessage{
		Text:        "📊 デイリーポートフォリオレポート",
		Attachments: attachments,
	}

	metadata := map[string]interface{}{
		"total_value":        summary.TotalValue,
		"total_cost":         summary.TotalCost,
		"total_gain":         summary.TotalGain,
		"total_gain_percent": summary.TotalGainPercent,
		"holdings_count":     len(summary.Holdings),
	}

	return s.sendSlackMessageWithLog(context.Background(), msg, "comprehensive_report", metadata)
}

// SetLogRepository sets the notification log repository
func (s *SlackNotifier) SetLogRepository(logRepo repository.NotificationLogRepository) {
	s.logRepo = logRepo
}

func (s *SlackNotifier) sendSlackMessage(msg SlackMessage) error {
	return s.sendSlackMessageWithLog(context.Background(), msg, "generic", nil)
}

// sendSlackMessageWithLog sends a Slack message and logs the transmission
func (s *SlackNotifier) sendSlackMessageWithLog(ctx context.Context, msg SlackMessage, notificationType string, metadata map[string]interface{}) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create notification log entry if repository is available
	var logID int64
	if s.logRepo != nil {
		metadataJSON, _ := json.Marshal(metadata)
		log := &repository.NotificationLog{
			NotificationType: notificationType,
			Status:           "pending",
			Message:          sql.NullString{String: msg.Text, Valid: true},
			Metadata:         metadataJSON,
			Attempts:         0,
		}
		if err := s.logRepo.Create(ctx, log); err != nil {
			logrus.Warnf("Failed to create notification log: %v", err)
		} else {
			logID = log.ID
		}
	}

	var lastErr error
	var attempts int
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		attempts = attempt + 1
		if attempt > 0 {
			logrus.Warnf("Retrying Slack notification (attempt %d/%d)", attempt, s.maxRetries)
			time.Sleep(s.retryDelay)
		}

		req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("User-Agent", "Stock-Automation/1.0")

		startTime := time.Now()
		resp, err := s.client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			lastErr = fmt.Errorf("failed to send message: %w", err)
			logrus.WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"error":    err,
				"duration": duration,
			}).Error("Failed to send Slack notification")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("slack API returned status code: %d", resp.StatusCode)
			logrus.WithFields(logrus.Fields{
				"attempt":     attempt + 1,
				"status_code": resp.StatusCode,
				"duration":    duration,
			}).Error("Slack API returned non-OK status")
			continue
		}

		// Success
		logrus.WithFields(logrus.Fields{
			"attempt":  attempt + 1,
			"duration": duration,
			"type":     notificationType,
		}).Info("Successfully sent Slack notification")

		// Update log entry with success
		if s.logRepo != nil && logID > 0 {
			now := time.Now()
			if err := s.logRepo.UpdateStatus(ctx, logID, "sent", nil, &now); err != nil {
				logrus.Warnf("Failed to update notification log: %v", err)
			}
		}

		return nil
	}

	// Update log entry with failure
	if s.logRepo != nil && logID > 0 {
		errMsg := lastErr.Error()
		if err := s.logRepo.UpdateStatus(ctx, logID, "failed", &errMsg, nil); err != nil {
			logrus.Warnf("Failed to update notification log: %v", err)
		}
	}

	return fmt.Errorf("failed to send Slack notification after %d attempts: %w", attempts, lastErr)
}
