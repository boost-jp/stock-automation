package interfaces

import (
	"context"
	"time"

	"github.com/boost-jp/stock-automation/app/infrastructure/alert"
	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

// DataScheduler manages scheduled tasks for the application
type DataScheduler struct {
	collectorUseCase   *usecase.CollectDataUseCase
	reporterUseCase    *usecase.PortfolioReportUseCase
	scheduler          *gocron.Scheduler
	recoveryMiddleware *alert.RecoveryMiddleware
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

// SetRecoveryMiddleware sets the recovery middleware for the scheduler
func (ds *DataScheduler) SetRecoveryMiddleware(middleware *alert.RecoveryMiddleware) {
	ds.recoveryMiddleware = middleware
}

// executeWithRecovery executes a function with panic recovery if middleware is available
func (ds *DataScheduler) executeWithRecovery(ctx context.Context, operation string, fn func() error) {
	if ds.recoveryMiddleware != nil {
		ds.recoveryMiddleware.WrapOperation(ctx, operation, fn)
	} else {
		if err := fn(); err != nil {
			logrus.WithError(err).Errorf("Error in %s", operation)
		}
	}
}

// StartScheduledCollection starts all scheduled tasks
func (ds *DataScheduler) StartScheduledCollection() {
	ctx := context.Background()

	// Every 5 minutes: Update prices (only during market hours)
	ds.scheduler.Every(5).Minutes().Do(func() {
		ds.executeWithRecovery(ctx, "UpdatePrices", func() error {
			if isMarketOpen() {
				return ds.collectorUseCase.UpdateAllPrices(ctx)
			}
			return nil
		})
	})

	// Every 30 minutes: Update configurations
	ds.scheduler.Every(30).Minutes().Do(func() {
		ds.executeWithRecovery(ctx, "UpdateWatchList", func() error {
			return ds.collectorUseCase.UpdateWatchList(ctx)
		})

		ds.executeWithRecovery(ctx, "UpdatePortfolio", func() error {
			return ds.collectorUseCase.UpdatePortfolio(ctx)
		})
	})

	// Daily at 8:00 AM JST: Send daily report
	ds.scheduler.Every(1).Day().At("08:00").Do(func() {
		ds.executeWithRecovery(ctx, "SendDailyReport", func() error {
			return ds.reporterUseCase.GenerateAndSendDailyReport(ctx)
		})
	})

	// Daily at 2:00 AM JST: Cleanup old data
	ds.scheduler.Every(1).Day().At("02:00").Do(func() {
		ds.executeWithRecovery(ctx, "CleanupOldData", func() error {
			return ds.collectorUseCase.CleanupOldData(ctx, 365)
		})
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
