package api

import (
	"time"

	"github.com/boost-jp/stock-automation/app/notification"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type DataScheduler struct {
	collector *DataCollector
	scheduler *gocron.Scheduler
	reporter  *DailyReporter
}

func NewDataScheduler(collector *DataCollector, notifier *notification.SlackNotifier) *DataScheduler {
	s := gocron.NewScheduler(time.UTC)
	reporter := NewDailyReporter(collector.db, notifier)

	return &DataScheduler{
		collector: collector,
		scheduler: s,
		reporter:  reporter,
	}
}

func (ds *DataScheduler) StartScheduledCollection() {
	// 5分毎の価格更新（市場時間中のみ）
	ds.scheduler.Every(5).Minutes().Do(func() {
		if ds.collector.IsMarketOpen() {
			if err := ds.collector.UpdateAllPrices(); err != nil {
				logrus.Error("Failed to update prices:", err)
			}
		}
	})

	// 30分毎の設定更新
	ds.scheduler.Every(30).Minutes().Do(func() {
		if err := ds.collector.UpdateWatchList(); err != nil {
			logrus.Error("Failed to update watch list:", err)
		}

		if err := ds.collector.UpdatePortfolio(); err != nil {
			logrus.Error("Failed to update portfolio:", err)
		}
	})

	// 毎日朝8時の日次レポート
	ds.scheduler.Every(1).Day().At("08:00").Do(func() {
		if err := ds.reporter.GenerateAndSendDailyReport(); err != nil {
			logrus.Error("Failed to send daily report:", err)
		}
	})

	// 毎日深夜のデータクリーンアップ
	ds.scheduler.Every(1).Day().At("02:00").Do(func() {
		if err := ds.collector.db.CleanupOldData(365); err != nil {
			logrus.Error("Failed to cleanup old data:", err)
		}
	})

	ds.scheduler.StartAsync()
	logrus.Info("Data collection scheduler started")
}

func (ds *DataScheduler) Stop() {
	ds.scheduler.Stop()
	logrus.Info("Data collection scheduler stopped")
}
