package main

import (
	"log"

	"github.com/boost-jp/stock-automation/app/api"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/usecase"
)

func main() {
	log.Println("📊 日次レポーター機能テスト開始")

	// データベース接続
	config := database.DefaultDatabaseConfig()
	connMgr, err := database.NewConnectionManager(config)
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
	}
	defer connMgr.Close()

	// Repository層初期化
	db := connMgr.GetDB()
	stockRepo := repository.NewStockRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	// 外部サービス初期化
	notifier := notification.NewSlackNotifier()
	stockClient := client.NewYahooFinanceClient()

	// UseCase初期化
	portfolioReportUseCase := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notifier)

	// DailyReporter初期化
	reporter := api.NewDailyReporter(portfolioReportUseCase)

	// 1. ポートフォリオ統計取得テスト
	log.Println("\n📈 ポートフォリオ統計取得テスト")

	statistics, err := reporter.GetPortfolioStatistics()
	if err != nil {
		log.Printf("❌ 統計取得エラー: %v", err)
	} else {
		log.Printf("✅ 統計取得成功:")
		log.Printf("   総資産: ¥%.0f", statistics.TotalValue)
		log.Printf("   損益: ¥%.0f (%.2f%%)", statistics.TotalGain, statistics.TotalGainPercent)
		log.Printf("   保有銘柄数: %d", len(statistics.Holdings))
	}

	// 2. 包括的レポート生成テスト
	log.Println("\n📋 包括的レポート生成テスト")

	report, err := reporter.GenerateComprehensiveDailyReport()
	if err != nil {
		log.Printf("❌ レポート生成エラー: %v", err)
	} else {
		log.Println("✅ レポート生成成功:")
		log.Println("=====================================")
		log.Println(report)
		log.Println("=====================================")
	}

	// 3. 基本レポート生成・送信テスト（実際には送信しない）
	log.Println("\n📤 基本レポート生成・送信テスト")

	err = reporter.GenerateAndSendDailyReport()
	if err != nil {
		log.Printf("❌ 基本レポート送信エラー: %v", err)
	} else {
		log.Println("✅ 基本レポート処理成功")
	}

	log.Println("\n🎉 日次レポーター機能テスト完了")
}
