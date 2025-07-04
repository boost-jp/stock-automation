package main

import (
	"fmt"
	"log"
	"os"

	"stock-automation/internal/api"
	"stock-automation/internal/database"
	"stock-automation/internal/notification"
)

func main() {
	// Slack Webhook URLを設定
	webhookURL := "https://hooks.slack.com/services/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"
	os.Setenv("SLACK_WEBHOOK_URL", webhookURL)

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 通知サービス初期化
	notifier := notification.NewSlackNotifier()

	// 日次レポーター初期化
	reporter := api.NewDailyReporter(db, notifier)

	fmt.Println("Testing daily report generation...")

	// 日次レポートを生成・送信
	if err := reporter.GenerateAndSendDailyReport(); err != nil {
		log.Fatal("Failed to generate daily report:", err)
	}

	fmt.Println("✅ Daily report sent successfully!")

	// 詳細ポートフォリオ分析も送信
	fmt.Println("\nTesting detailed portfolio analysis...")
	if err := reporter.SendPortfolioAnalysis(); err != nil {
		log.Fatal("Failed to send portfolio analysis:", err)
	}

	fmt.Println("✅ Portfolio analysis sent successfully!")
	fmt.Println("\n🎉 Daily reporting system is working correctly!")
}
