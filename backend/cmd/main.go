package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boost-jp/stock-automation/app/api"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/repository"
	"github.com/sirupsen/logrus"
)

func main() {
	// ログ設定
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false,
	})

	// データベース接続
	config := database.DefaultDatabaseConfig()
	connMgr, err := database.NewConnectionManager(config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer connMgr.Close()

	// Repository層初期化
	txMgr := repository.NewTransactionManager(connMgr.GetDB())
	repos := txMgr.GetRepositories()

	// 通知サービス初期化
	notifier := notification.NewSlackNotifier()

	// データコレクター初期化
	collector := api.NewDataCollector(repos)

	// 監視銘柄とポートフォリオの初期読み込み
	if err := collector.UpdateWatchList(); err != nil {
		logrus.Error("Failed to initialize watch list:", err)
	}

	if err := collector.UpdatePortfolio(); err != nil {
		logrus.Error("Failed to initialize portfolio:", err)
	}

	// スケジューラー開始
	scheduler := api.NewDataScheduler(collector, notifier)
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
