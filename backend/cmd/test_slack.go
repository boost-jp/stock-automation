package main

import (
	"fmt"
	"log"
	"os"

	"stock-automation/internal/notification"
)

func main() {
	// Slack Webhook URLを環境変数から取得
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "https://hooks.slack.com/services/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"
		os.Setenv("SLACK_WEBHOOK_URL", webhookURL)
	}

	// Slack通知サービスの初期化
	notifier := notification.NewSlackNotifier()

	// テストメッセージの送信
	fmt.Println("Testing Slack notification...")
	if err := notifier.SendMessage("🧪 株価自動化システムのテスト通知です"); err != nil {
		log.Fatal("Failed to send test message:", err)
	}
	fmt.Println("✅ Test message sent successfully!")

	// 株価アラートのテスト
	fmt.Println("\nTesting stock alert notification...")
	if err := notifier.SendStockAlert("7203", "トヨタ自動車", 2484.50, 2500.00, "buy"); err != nil {
		log.Fatal("Failed to send stock alert:", err)
	}
	fmt.Println("✅ Stock alert sent successfully!")

	// 日次レポートのテスト
	fmt.Println("\nTesting daily report notification...")
	if err := notifier.SendDailyReport(1000000.00, 50000.00, 5.25); err != nil {
		log.Fatal("Failed to send daily report:", err)
	}
	fmt.Println("✅ Daily report sent successfully!")

	fmt.Println("\n🎉 All Slack notifications are working correctly!")
}