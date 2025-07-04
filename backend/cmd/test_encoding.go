package main

import (
	"fmt"
	"log"
	"os"

	"stock-automation/internal/api"
	"stock-automation/internal/database"
	"stock-automation/internal/notification"

	"github.com/sirupsen/logrus"
)

func main() {
	// Slack Webhook URLを設定
	webhookURL := "https://hooks.slack.com/services/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"
	os.Setenv("SLACK_WEBHOOK_URL", webhookURL)

	// ログレベル設定
	logrus.SetLevel(logrus.InfoLevel)

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// データベース初期化
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("🧪 文字エンコーディングテスト開始")

	// データコレクター初期化
	collector := api.NewDataCollector(db)
	
	// 監視銘柄とポートフォリオの初期読み込み
	if err := collector.UpdateWatchList(); err != nil {
		logrus.Error("Failed to update watch list:", err)
	} else {
		logrus.Info("監視銘柄リストを更新しました")
	}
	
	if err := collector.UpdatePortfolio(); err != nil {
		logrus.Error("Failed to update portfolio:", err)
	} else {
		logrus.Info("ポートフォリオを更新しました")
	}

	// 通知サービス初期化
	notifier := notification.NewSlackNotifier()

	// 日本語メッセージのテスト
	fmt.Println("\n📱 Slack日本語メッセージテスト")
	if err := notifier.SendMessage("🇯🇵 日本語文字化けテスト\n漢字・ひらがな・カタカナ・記号が正しく表示されるかのテストです。\n株価情報: トヨタ自動車 ¥2,484"); err != nil {
		log.Fatal("Failed to send Japanese test message:", err)
	}
	fmt.Println("✅ 日本語メッセージ送信完了")

	// 株価データ取得テスト
	fmt.Println("\n📊 株価データ取得テスト")
	yahooClient := api.NewYahooFinanceClient()
	
	// トヨタ自動車のデータ取得
	price, err := yahooClient.GetCurrentPrice("7203")
	if err != nil {
		log.Fatal("Failed to get price:", err)
	}
	
	price.Name = "トヨタ自動車"
	if err := db.SaveStockPrice(price); err != nil {
		log.Fatal("Failed to save price:", err)
	}
	
	logrus.Infof("株価データ保存: %s ¥%.2f", price.Name, price.Price)
	fmt.Printf("✅ %s の株価 ¥%.2f を保存しました\n", price.Name, price.Price)

	// データベースから日本語データ読み込みテスト
	fmt.Println("\n💾 データベース日本語読み込みテスト")
	latestPrice, err := db.GetLatestPrice("7203")
	if err != nil {
		log.Fatal("Failed to get latest price:", err)
	}
	
	fmt.Printf("✅ データベースから読み込み: %s ¥%.2f\n", latestPrice.Name, latestPrice.Price)
	logrus.Infof("データベース読み込み成功: %s", latestPrice.Name)

	// 株価アラート送信（日本語）
	fmt.Println("\n🔔 日本語株価アラートテスト")
	if err := notifier.SendStockAlert("7203", "トヨタ自動車", latestPrice.Price, 2500.00, "買い推奨"); err != nil {
		log.Fatal("Failed to send stock alert:", err)
	}
	fmt.Println("✅ 日本語株価アラート送信完了")

	fmt.Println("\n🎉 すべての文字エンコーディングテストが完了しました！")
	logrus.Info("文字エンコーディングテスト完了")
}