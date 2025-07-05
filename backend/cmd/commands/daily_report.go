package commands

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/boost-jp/stock-automation/app/api"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/usecase"
)

// RunDailyReportTester runs the daily report tester command.
func RunDailyReportTester(connMgr database.ConnectionManager, args []string) {
	// Command line flags
	reportCmd := flag.NewFlagSet("daily-report", flag.ExitOnError)
	var (
		testMode      = reportCmd.Bool("test", false, "Run in test mode (don't send actual notifications)")
		sendToSlack   = reportCmd.Bool("send", false, "Actually send report to Slack")
		comprehensive = reportCmd.Bool("comprehensive", false, "Generate comprehensive report")
	)

	reportCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: stock-automation daily-report [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Generate and send daily portfolio report\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		reportCmd.PrintDefaults()
	}

	if err := reportCmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	log.Println("📊 日次レポーター機能開始")

	// Repository層初期化
	db := connMgr.GetDB()
	stockRepo := repository.NewStockRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	// 外部サービス初期化
	var notifier notification.NotificationService
	if *testMode || !*sendToSlack {
		// テストモードまたは送信フラグが無い場合はダミーノーティファイアを使用
		notifier = &dummyNotifier{}
	} else {
		notifier = notification.NewSlackNotifier()
	}

	stockClient := client.NewYahooFinanceClient()

	// UseCase初期化
	portfolioReportUseCase := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notifier)

	// DailyReporter初期化
	reporter := api.NewDailyReporter(portfolioReportUseCase)

	if *comprehensive {
		// 包括的レポート生成
		log.Println("\n📋 包括的レポート生成中...")

		report, err := reporter.GenerateComprehensiveDailyReport()
		if err != nil {
			log.Fatalf("❌ レポート生成エラー: %v", err)
		}

		log.Println("✅ レポート生成成功:")
		log.Println("=====================================")
		fmt.Println(report)
		log.Println("=====================================")

		if *sendToSlack && !*testMode {
			// Slackに送信
			if err := notifier.SendMessage(report); err != nil {
				log.Printf("❌ Slack送信エラー: %v", err)
			} else {
				log.Println("✅ Slackに送信しました")
			}
		}
	} else {
		// 基本レポート送信
		err := reporter.GenerateAndSendDailyReport()
		if err != nil {
			log.Fatalf("❌ レポート送信エラー: %v", err)
		}
		log.Println("✅ 基本レポート処理成功")
	}

	// ポートフォリオ統計も表示
	log.Println("\n📈 ポートフォリオ統計:")
	statistics, err := reporter.GetPortfolioStatistics()
	if err != nil {
		log.Printf("❌ 統計取得エラー: %v", err)
	} else {
		log.Printf("   総資産: ¥%.0f", statistics.TotalValue)
		log.Printf("   損益: ¥%.0f (%.2f%%)", statistics.TotalGain, statistics.TotalGainPercent)
		log.Printf("   保有銘柄数: %d", len(statistics.Holdings))
	}

	log.Println("\n🎉 日次レポーター機能完了")
}

// dummyNotifier is a test notifier that logs messages instead of sending them.
type dummyNotifier struct{}

func (d *dummyNotifier) SendMessage(message string) error {
	log.Printf("[DummyNotifier] Message would be sent:\n%s", message)
	return nil
}

func (d *dummyNotifier) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	log.Printf("[DummyNotifier] Stock alert would be sent: %s (%s) - Current: ¥%.2f, Target: ¥%.2f, Type: %s",
		stockName, stockCode, currentPrice, targetPrice, alertType)
	return nil
}

func (d *dummyNotifier) SendDailyReport(totalValue, totalGain float64, gainPercent float64) error {
	log.Printf("[DummyNotifier] Daily report would be sent: Total: ¥%.0f, Gain: ¥%.0f (%.2f%%)",
		totalValue, totalGain, gainPercent)
	return nil
}

