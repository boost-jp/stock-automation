package interfaces

import (
	"context"
	"time"

	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

// DataScheduler manages scheduled tasks for the application
type DataScheduler struct {
	collectorUseCase *usecase.CollectDataUseCase
	reporterUseCase  *usecase.PortfolioReportUseCase
	scheduler        *gocron.Scheduler
}

// NewDataScheduler creates a new data scheduler
func NewDataScheduler(
	collectorUseCase *usecase.CollectDataUseCase,
	reporterUseCase *usecase.PortfolioReportUseCase,
) *DataScheduler {
	s := gocron.NewScheduler(time.FixedZone("JST", 9*60*60))

	return &DataScheduler{
		collectorUseCase: collectorUseCase,
		reporterUseCase:  reporterUseCase,
		scheduler:        s,
	}
}

// StartScheduledCollection starts all scheduled tasks
func (ds *DataScheduler) StartScheduledCollection() {
	ctx := context.Background()

	// Every 5 minutes: Update prices (only during market hours)
	ds.scheduler.Every(5).Minutes().Do(func() {
		if isMarketOpen() {
			if err := ds.collectorUseCase.UpdateAllPrices(ctx); err != nil {
				logrus.Error("Failed to update prices:", err)
			}
		}
	})

	// Every 30 minutes: Update configurations
	ds.scheduler.Every(30).Minutes().Do(func() {
		if err := ds.collectorUseCase.UpdateWatchList(ctx); err != nil {
			logrus.Error("Failed to update watch list:", err)
		}

		if err := ds.collectorUseCase.UpdatePortfolio(ctx); err != nil {
			logrus.Error("Failed to update portfolio:", err)
		}
	})

	// Daily at 8:00 AM JST: Send daily report
	ds.scheduler.Every(1).Day().At("08:00").Do(func() {
		if err := ds.reporterUseCase.GenerateAndSendDailyReport(ctx); err != nil {
			logrus.Error("Failed to send daily report:", err)
		}
	})

	// Daily at 2:00 AM JST: Cleanup old data
	ds.scheduler.Every(1).Day().At("02:00").Do(func() {
		if err := ds.collectorUseCase.CleanupOldData(ctx, 365); err != nil {
			logrus.Error("Failed to cleanup old data:", err)
		}
	})

	ds.scheduler.StartAsync()
	logrus.Info("Data collection scheduler started")
}

// Stop stops all scheduled tasks
func (ds *DataScheduler) Stop() {
	ds.scheduler.Stop()
	logrus.Info("Data collection scheduler stopped")
}

// isMarketOpen checks if the Japanese stock market is currently open
func isMarketOpen() bool {
	now := time.Now().In(time.FixedZone("JST", 9*60*60))
	weekday := now.Weekday()

	// Market is closed on weekends
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Market hours: 9:00 AM - 3:00 PM JST (with lunch break 11:30 AM - 12:30 PM)
	hour := now.Hour()
	minute := now.Minute()
	currentMinutes := hour*60 + minute

	morningOpen := 9 * 60       // 9:00 AM
	morningClose := 11*60 + 30  // 11:30 AM
	afternoonOpen := 12*60 + 30 // 12:30 PM
	afternoonClose := 15 * 60   // 3:00 PM

	return (currentMinutes >= morningOpen && currentMinutes < morningClose) ||
		(currentMinutes >= afternoonOpen && currentMinutes < afternoonClose)
}
