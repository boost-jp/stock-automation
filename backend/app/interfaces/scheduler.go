package interfaces

import (
	"context"
	"time"

	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

// DataScheduler manages scheduled tasks for the application
type DataScheduler struct {
	collectorUseCase *usecase.CollectDataUseCase
	reporterUseCase  *usecase.PortfolioReportUseCase
	scheduler        *gocron.Scheduler
	logRepo          repository.SchedulerLogRepository
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
			ds.executeWithLogging(ctx, "update_prices", func(ctx context.Context) error {
				return ds.collectorUseCase.UpdateAllPrices(ctx)
			})
		}
	})

	// Every 30 minutes: Update configurations
	ds.scheduler.Every(30).Minutes().Do(func() {
		ds.executeWithLogging(ctx, "update_configurations", func(ctx context.Context) error {
			if err := ds.collectorUseCase.UpdateWatchList(ctx); err != nil {
				return err
			}
			return ds.collectorUseCase.UpdatePortfolio(ctx)
		})
	})

	// Weekdays at 18:00 JST: Send daily report
	ds.scheduler.Monday().Tuesday().Wednesday().Thursday().Friday().At("18:00").Do(func() {
		ds.executeWithLogging(ctx, "daily_report", func(ctx context.Context) error {
			return ds.reporterUseCase.GenerateAndSendDailyReport(ctx)
		})
	})

	// Daily at 2:00 AM JST: Cleanup old data
	ds.scheduler.Every(1).Day().At("02:00").Do(func() {
		ds.executeWithLogging(ctx, "cleanup_old_data", func(ctx context.Context) error {
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

// SetLogRepository sets the scheduler log repository
func (ds *DataScheduler) SetLogRepository(logRepo repository.SchedulerLogRepository) {
	ds.logRepo = logRepo
}

// executeWithLogging executes a task with logging
func (ds *DataScheduler) executeWithLogging(ctx context.Context, taskName string, task func(context.Context) error) {
	start := time.Now()
	var logID int64

	// Start logging if repository is available
	if ds.logRepo != nil {
		id, err := ds.logRepo.StartTask(ctx, taskName)
		if err != nil {
			logrus.Warnf("Failed to start task log for %s: %v", taskName, err)
		} else {
			logID = id
		}
	}

	logrus.Infof("Starting scheduled task: %s", taskName)

	// Execute the task
	err := task(ctx)
	duration := time.Since(start)

	if err != nil {
		logrus.Errorf("Failed to execute %s: %v (duration: %v)", taskName, err, duration)
		if ds.logRepo != nil && logID > 0 {
			if logErr := ds.logRepo.FailTask(ctx, logID, duration, err); logErr != nil {
				logrus.Warnf("Failed to update task log for %s: %v", taskName, logErr)
			}
		}
	} else {
		logrus.Infof("Completed scheduled task: %s (duration: %v)", taskName, duration)
		if ds.logRepo != nil && logID > 0 {
			if logErr := ds.logRepo.CompleteTask(ctx, logID, duration); logErr != nil {
				logrus.Warnf("Failed to update task log for %s: %v", taskName, logErr)
			}
		}
	}
}
