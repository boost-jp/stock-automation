package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type SlackNotifier struct {
	webhookURL string
	client     *http.Client
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

	return s.sendSlackMessage(msg)
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

	return s.sendSlackMessage(msg)
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

	return s.sendSlackMessage(msg)
}

func (s *SlackNotifier) sendSlackMessage(msg SlackMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Stock-Automation/1.0")

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
