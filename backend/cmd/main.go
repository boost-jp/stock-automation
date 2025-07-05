package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boost-jp/stock-automation/app/api"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/boost-jp/stock-automation/cmd/commands"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// グローバルフラグ
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)

	// サブコマンドの定義
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Stock Automation System v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: stock-automation [global flags] <command> [command flags]\n\n")
		fmt.Fprintf(os.Stderr, "Global flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  server              Start the main automation server\n")
		fmt.Fprintf(os.Stderr, "  watchlist           Manage watch list items\n")
		fmt.Fprintf(os.Stderr, "  bulk-collect        Collect historical data in bulk\n")
		fmt.Fprintf(os.Stderr, "  daily-report        Send daily portfolio report\n")
		fmt.Fprintf(os.Stderr, "  test-yahoo          Test Yahoo Finance API\n")
		fmt.Fprintf(os.Stderr, "  add-portfolio       Add sample portfolio data\n")
		fmt.Fprintf(os.Stderr, "\nRun 'stock-automation <command> -h' for more information on a command.\n")
	}

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("Stock Automation System\n")
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Built: %s\n", date)
		os.Exit(0)
	}

	// ログレベル設定
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %s", *logLevel)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false,
	})

	// サブコマンドの取得
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]

	// データベース接続（すべてのコマンドで必要）
	config := database.DefaultDatabaseConfig()
	connMgr, err := database.NewConnectionManager(config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer connMgr.Close()

	// サブコマンドの実行
	switch command {
	case "server":
		runServer(connMgr, args[1:])
	case "watchlist":
		commands.RunWatchListManager(connMgr, args[1:])
	case "bulk-collect":
		commands.RunBulkDataCollector(connMgr, args[1:])
	case "daily-report":
		commands.RunDailyReportTester(connMgr, args[1:])
	case "test-yahoo":
		commands.RunTestYahooAPI(args[1:])
	case "add-portfolio":
		commands.RunAddSamplePortfolio(connMgr, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func runServer(connMgr database.ConnectionManager, args []string) {
	// サーバー用のフラグ
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: stock-automation server [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Start the main automation server\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		serverCmd.PrintDefaults()
	}

	if err := serverCmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Repository層初期化 (個別のrepositoryを使用)
	db := connMgr.GetDB()
	stockRepo := repository.NewStockRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	// 外部サービス初期化
	notifier := notification.NewSlackNotifier()
	stockClient := client.NewYahooFinanceClient()

	// UseCase層初期化
	collectDataUseCase := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	portfolioReportUseCase := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notifier)

	// API層初期化 (UseCaseをラップ)
	collector := api.NewDataCollector(collectDataUseCase)
	reporter := api.NewDailyReporter(portfolioReportUseCase)

	// 監視銘柄とポートフォリオの初期読み込み
	if err := collector.UpdateWatchList(); err != nil {
		logrus.Error("Failed to initialize watch list:", err)
	}

	if err := collector.UpdatePortfolio(); err != nil {
		logrus.Error("Failed to initialize portfolio:", err)
	}

	// スケジューラー開始
	scheduler := api.NewDataScheduler(collector, reporter)
	scheduler.StartScheduledCollection()

	if err := notifier.SendMessage("📈 Stock Automation System Started"); err != nil {
		logrus.Error("Failed to send startup notification:", err)
	}

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logrus.Info("Received shutdown signal")
		scheduler.Stop()

		// 終了通知
		if err := notifier.SendMessage("🔴 Stock Automation System Stopped"); err != nil {
			logrus.Error("Failed to send shutdown notification:", err)
		}

		// グレースフルシャットダウン
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		<-shutdownCtx.Done()
		logrus.Info("Application stopped")
	}
}

