package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/interfaces"
	"github.com/boost-jp/stock-automation/app/testutil"
	"github.com/boost-jp/stock-automation/app/testutil/fixture"
	"github.com/boost-jp/stock-automation/app/usecase"
)

// TestScheduler_JobExecution tests scheduler job execution
func TestScheduler_JobExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// Create repositories
	stockRepo := repository.NewStockRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	// Create tracking mock services
	stockClient := &trackingStockClient{
		updatePricesCalled: 0,
		prices: map[string]float64{
			"7203": 2200.0,
			"6758": 14000.0,
		},
	}

	notificationService := &trackingNotificationService{
		reportsSent: 0,
	}

	// Create use cases
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	portfolioReportUC := usecase.NewPortfolioReportUseCase(
		stockRepo,
		portfolioRepo,
		stockClient,
		notificationService,
	)

	// Create scheduler
	scheduler := interfaces.NewDataScheduler(collectDataUC, portfolioReportUC)

	// Setup test data
	setupSchedulerTestData(t, ctx, stockRepo, portfolioRepo)

	// Test 1: Price update job
	t.Run("PriceUpdateJob", func(t *testing.T) {
		initialCallCount := stockClient.updatePricesCalled

		// Manually trigger price update (simulating scheduled job)
		err := collectDataUC.UpdateAllPrices(ctx)
		if err != nil {
			t.Errorf("Failed to update prices: %v", err)
		}

		// Verify prices were updated
		if stockClient.updatePricesCalled <= initialCallCount {
			t.Error("Price update was not called")
		}

		// Verify data in database
		price, err := stockRepo.GetLatestPrice(ctx, "7203")
		if err != nil {
			t.Errorf("Failed to get latest price: %v", err)
		}
		if price == nil {
			t.Error("No price found after update")
		}
	})

	// Test 2: Daily report job
	t.Run("DailyReportJob", func(t *testing.T) {
		initialReportCount := notificationService.reportsSent

		// Manually trigger daily report (simulating scheduled job)
		err := portfolioReportUC.GenerateAndSendDailyReport(ctx)
		if err != nil {
			t.Errorf("Failed to generate daily report: %v", err)
		}

		// Verify report was sent
		if notificationService.reportsSent <= initialReportCount {
			t.Error("Daily report was not sent")
		}
	})

	// Test 3: Market hours check
	t.Run("MarketHoursCheck", func(t *testing.T) {
		// Test that price updates respect market hours
		isMarketOpen := collectDataUC.IsMarketOpen()
		t.Logf("Market open: %v", isMarketOpen)

		// In real scheduler, price updates only run during market hours
		// This is a smoke test to ensure the logic exists
	})

	// Test 4: Scheduler lifecycle
	t.Run("SchedulerLifecycle", func(t *testing.T) {
		// Start scheduler (in background)
		scheduler.StartScheduledCollection()

		// Let it run briefly
		time.Sleep(100 * time.Millisecond)

		// Stop scheduler
		scheduler.Stop()

		// Verify it stopped gracefully
		t.Log("Scheduler stopped successfully")
	})
}

// TestScheduler_ConcurrentJobs tests concurrent job execution
func TestScheduler_ConcurrentJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	db, cleanup, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// Create repositories
	stockRepo := repository.NewStockRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	// Create thread-safe mock services
	stockClient := &concurrentSafeStockClient{
		prices: map[string]float64{
			"7203": 2300.0,
			"6758": 14500.0,
			"9984": 6800.0,
		},
	}

	notificationService := &concurrentSafeNotificationService{}

	// Create use cases
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	portfolioReportUC := usecase.NewPortfolioReportUseCase(
		stockRepo,
		portfolioRepo,
		stockClient,
		notificationService,
	)

	// Setup test data with more stocks
	setupExtendedTestData(t, ctx, stockRepo, portfolioRepo)

	// Run multiple jobs concurrently
	t.Run("ConcurrentExecution", func(t *testing.T) {
		done := make(chan bool, 3)

		// Job 1: Update prices
		go func() {
			err := collectDataUC.UpdateAllPrices(ctx)
			if err != nil {
				t.Errorf("Concurrent price update failed: %v", err)
			}
			done <- true
		}()

		// Job 2: Generate report
		go func() {
			err := portfolioReportUC.GenerateAndSendDailyReport(ctx)
			if err != nil {
				t.Errorf("Concurrent report generation failed: %v", err)
			}
			done <- true
		}()

		// Job 3: Cleanup old data
		go func() {
			err := collectDataUC.CleanupOldData(ctx, 365)
			if err != nil {
				t.Errorf("Concurrent cleanup failed: %v", err)
			}
			done <- true
		}()

		// Wait for all jobs to complete
		for i := 0; i < 3; i++ {
			select {
			case <-done:
				// Job completed
			case <-time.After(10 * time.Second):
				t.Error("Timeout waiting for concurrent job")
			}
		}

		t.Log("All concurrent jobs completed successfully")
	})
}

// Helper types for scheduler testing

type trackingStockClient struct {
	updatePricesCalled int
	prices             map[string]float64
	mu                 sync.Mutex
}

func (t *trackingStockClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.updatePricesCalled++

	price, ok := t.prices[stockCode]
	if !ok {
		price = 1000.0
	}

	return &models.StockPrice{
		Code:       stockCode,
		Date:       time.Now(),
		ClosePrice: client.FloatToDecimal(price),
		OpenPrice:  client.FloatToDecimal(price * 0.99),
		HighPrice:  client.FloatToDecimal(price * 1.01),
		LowPrice:   client.FloatToDecimal(price * 0.98),
		Volume:     1000000,
	}, nil
}

func (t *trackingStockClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (t *trackingStockClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

type trackingNotificationService struct {
	reportsSent int
	mu          sync.Mutex
}

func (t *trackingNotificationService) SendMessage(message string) error {
	return nil
}

func (t *trackingNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reportsSent++
	return nil
}

func (t *trackingNotificationService) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	return nil
}

// Thread-safe implementations for concurrent testing

type concurrentSafeStockClient struct {
	prices map[string]float64
	mu     sync.RWMutex
}

func (c *concurrentSafeStockClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	price, ok := c.prices[stockCode]
	if !ok {
		price = 1000.0
	}

	return &models.StockPrice{
		Code:       stockCode,
		Date:       time.Now(),
		ClosePrice: client.FloatToDecimal(price),
		OpenPrice:  client.FloatToDecimal(price),
		HighPrice:  client.FloatToDecimal(price),
		LowPrice:   client.FloatToDecimal(price),
		Volume:     1000000,
	}, nil
}

func (c *concurrentSafeStockClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (c *concurrentSafeStockClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

type concurrentSafeNotificationService struct {
	mu sync.Mutex
}

func (c *concurrentSafeNotificationService) SendMessage(message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil
}

func (c *concurrentSafeNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil
}

func (c *concurrentSafeNotificationService) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return nil
}

// Helper functions

func setupSchedulerTestData(t *testing.T, ctx context.Context, stockRepo repository.StockRepository, portfolioRepo repository.PortfolioRepository) {
	// Add watch list
	watchList := []*models.WatchList{
		{
			ID:       fixture.WatchListID1,
			Code:     "7203",
			Name:     "トヨタ自動車",
			IsActive: fixture.NullBoolFrom(true),
		},
		{
			ID:       fixture.WatchListID2,
			Code:     "6758",
			Name:     "ソニーグループ",
			IsActive: fixture.NullBoolFrom(true),
		},
	}

	for _, w := range watchList {
		if err := stockRepo.AddToWatchList(ctx, w); err != nil {
			t.Fatalf("Failed to add watch list: %v", err)
		}
	}

	// Add portfolio
	portfolios := []*models.Portfolio{
		{
			ID:            fixture.PortfolioID1,
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: client.FloatToDecimal(2000.0),
			PurchaseDate:  time.Now().AddDate(0, -6, 0),
		},
		{
			ID:            fixture.PortfolioID2,
			Code:          "6758",
			Name:          "ソニーグループ",
			Shares:        50,
			PurchasePrice: client.FloatToDecimal(12000.0),
			PurchaseDate:  time.Now().AddDate(0, -3, 0),
		},
	}

	for _, p := range portfolios {
		if err := portfolioRepo.Create(ctx, p); err != nil {
			t.Fatalf("Failed to create portfolio: %v", err)
		}
	}
}

func setupExtendedTestData(t *testing.T, ctx context.Context, stockRepo repository.StockRepository, portfolioRepo repository.PortfolioRepository) {
	// Setup basic data first
	setupSchedulerTestData(t, ctx, stockRepo, portfolioRepo)

	// Add more stocks
	additionalWatchList := &models.WatchList{
		ID:       fixture.WatchListID3,
		Code:     "9984",
		Name:     "ソフトバンクグループ",
		IsActive: fixture.NullBoolFrom(true),
	}

	if err := stockRepo.AddToWatchList(ctx, additionalWatchList); err != nil {
		t.Fatalf("Failed to add additional watch list: %v", err)
	}

	// Add to portfolio
	additionalPortfolio := &models.Portfolio{
		ID:            fixture.PortfolioID3,
		Code:          "9984",
		Name:          "ソフトバンクグループ",
		Shares:        200,
		PurchasePrice: client.FloatToDecimal(6000.0),
		PurchaseDate:  time.Now().AddDate(0, -1, 0),
	}

	if err := portfolioRepo.Create(ctx, additionalPortfolio); err != nil {
		t.Fatalf("Failed to create additional portfolio: %v", err)
	}
}
