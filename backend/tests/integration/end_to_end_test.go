package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/config"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/interfaces"
	"github.com/boost-jp/stock-automation/app/testutil"
	"github.com/boost-jp/stock-automation/app/testutil/fixture"
	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/sirupsen/logrus"
)

// TestEndToEnd_DataCollectionToReporting tests the complete flow from data collection to report generation
func TestEndToEnd_DataCollectionToReporting(t *testing.T) {
	// Skip if not in integration test mode
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

	// Create mock external services
	stockClient := &e2eMockStockDataClient{
		prices: map[string]float64{
			"7203": 2100.0,  // トヨタ
			"6758": 13500.0, // ソニー
		},
	}

	notificationService := &e2eMockNotificationService{}

	// Create use cases
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	portfolioReportUC := usecase.NewPortfolioReportUseCase(
		stockRepo,
		portfolioRepo,
		stockClient,
		notificationService,
	)

	// Step 1: Setup test data - Portfolio
	portfolios := []*models.Portfolio{
		{
			ID:            fixture.PortfolioID1,
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: client.FloatToDecimal(2000.0),
			PurchaseDate:  time.Now().AddDate(0, -6, 0),
			CreatedAt:     fixture.NullTimeFrom(time.Now()),
			UpdatedAt:     fixture.NullTimeFrom(time.Now()),
		},
		{
			ID:            fixture.PortfolioID2,
			Code:          "6758",
			Name:          "ソニーグループ",
			Shares:        50,
			PurchasePrice: client.FloatToDecimal(12000.0),
			PurchaseDate:  time.Now().AddDate(0, -3, 0),
			CreatedAt:     fixture.NullTimeFrom(time.Now()),
			UpdatedAt:     fixture.NullTimeFrom(time.Now()),
		},
	}

	for _, p := range portfolios {
		if err := portfolioRepo.Create(ctx, p); err != nil {
			t.Fatalf("Failed to create portfolio: %v", err)
		}
	}

	// Step 2: Setup test data - Watch List
	watchList := []*models.WatchList{
		{
			ID:        fixture.WatchListID1,
			Code:      "7203",
			Name:      "トヨタ自動車",
			IsActive:  fixture.NullBoolFrom(true),
			CreatedAt: fixture.NullTimeFrom(time.Now()),
			UpdatedAt: fixture.NullTimeFrom(time.Now()),
		},
		{
			ID:        fixture.WatchListID2,
			Code:      "6758",
			Name:      "ソニーグループ",
			IsActive:  fixture.NullBoolFrom(true),
			CreatedAt: fixture.NullTimeFrom(time.Now()),
			UpdatedAt: fixture.NullTimeFrom(time.Now()),
		},
	}

	for _, w := range watchList {
		if err := stockRepo.AddToWatchList(ctx, w); err != nil {
			t.Fatalf("Failed to add to watch list: %v", err)
		}
	}

	// Step 3: Collect current prices
	t.Run("DataCollection", func(t *testing.T) {
		err := collectDataUC.UpdateAllPrices(ctx)
		if err != nil {
			t.Errorf("Failed to update prices: %v", err)
		}

		// Verify prices were saved
		for code := range stockClient.prices {
			price, err := stockRepo.GetLatestPrice(ctx, code)
			if err != nil {
				t.Errorf("Failed to get latest price for %s: %v", code, err)
			}
			if price == nil {
				t.Errorf("No price found for %s", code)
			}
		}
	})

	// Step 4: Generate and send daily report
	t.Run("ReportGeneration", func(t *testing.T) {
		err := portfolioReportUC.GenerateAndSendDailyReport(ctx)
		if err != nil {
			t.Errorf("Failed to generate daily report: %v", err)
		}

		// Verify notification was sent
		if !notificationService.sendDailyReportCalled {
			t.Error("Daily report notification was not sent")
		}

		// Verify report content
		expectedTotalValue := 2100.0*100 + 13500.0*50 // 210,000 + 675,000 = 885,000
		if notificationService.lastTotalValue != expectedTotalValue {
			t.Errorf("Expected total value %f, got %f", expectedTotalValue, notificationService.lastTotalValue)
		}
	})

	// Step 5: Test comprehensive report
	t.Run("ComprehensiveReport", func(t *testing.T) {
		err := portfolioReportUC.SendComprehensiveDailyReport(ctx)
		if err != nil {
			t.Errorf("Failed to send comprehensive report: %v", err)
		}

		// Verify notification was sent
		if !notificationService.sendMessageCalled {
			t.Error("Comprehensive report notification was not sent")
		}

		// Verify report contains expected elements
		if notificationService.lastMessage == "" {
			t.Error("Report message is empty")
		}
	})
}

// TestEndToEnd_SchedulerIntegration tests the scheduler integration
func TestEndToEnd_SchedulerIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	_, cleanup, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         3306,
			User:         "test",
			Password:     "test",
			DatabaseName: "test_db",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			MaxLifetime:  time.Hour,
		},
		Yahoo: config.YahooConfig{
			BaseURL:       "https://test.example.com",
			Timeout:       30 * time.Second,
			RetryCount:    3,
			RetryWaitTime: time.Second,
			RetryMaxWait:  10 * time.Second,
			RateLimitRPS:  10,
		},
		Slack: config.SlackConfig{
			WebhookURL: "",
			Channel:    "#test",
			Username:   "Test Bot",
		},
	}

	// Create container with test configuration
	container, err := interfaces.NewContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	// Get scheduler
	scheduler := container.GetScheduler()

	// Start scheduler in test mode (we'll manually trigger jobs)
	t.Run("SchedulerJobs", func(t *testing.T) {
		// Test that scheduler can be started without errors
		// Note: We don't actually run scheduled jobs in tests
		if scheduler == nil {
			t.Error("Scheduler is nil")
		}

		// Verify scheduler has expected jobs configured
		// This is more of a smoke test
		logrus.Info("Scheduler created successfully")
	})
}

// TestEndToEnd_ErrorHandling tests error handling across the system
func TestEndToEnd_ErrorHandling(t *testing.T) {
	// Skip if not in integration test mode
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

	// Create mock services that return errors
	stockClient := &e2eMockStockDataClient{
		shouldError: true,
		errorMsg:    "API rate limit exceeded",
	}

	notificationService := &e2eMockNotificationService{
		shouldError: true,
		errorMsg:    "Slack webhook failed",
	}

	// Create use cases
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	portfolioReportUC := usecase.NewPortfolioReportUseCase(
		stockRepo,
		portfolioRepo,
		stockClient,
		notificationService,
	)

	// Test error handling in data collection
	t.Run("DataCollectionError", func(t *testing.T) {
		// Setup watch list
		watchList := &models.WatchList{
			ID:       fixture.WatchListID1,
			Code:     "7203",
			Name:     "トヨタ自動車",
			IsActive: fixture.NullBoolFrom(true),
		}
		stockRepo.AddToWatchList(ctx, watchList)

		// This should handle errors gracefully
		err := collectDataUC.UpdateAllPrices(ctx)
		// UpdateAllPrices doesn't return error for individual failures
		if err != nil {
			t.Logf("UpdateAllPrices returned error as expected: %v", err)
		}
	})

	// Test error handling in report generation
	t.Run("ReportGenerationError", func(t *testing.T) {
		// Setup portfolio
		portfolio := &models.Portfolio{
			ID:            fixture.PortfolioID1,
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: client.FloatToDecimal(2000.0),
			PurchaseDate:  time.Now().AddDate(0, -6, 0),
		}
		portfolioRepo.Create(ctx, portfolio)

		// This should handle notification errors gracefully
		err := portfolioReportUC.GenerateAndSendDailyReport(ctx)
		if err == nil {
			t.Error("Expected error from notification service, got nil")
		}
	})
}

// Enhanced mock for comprehensive testing
type e2eMockStockDataClient struct {
	prices      map[string]float64
	shouldError bool
	errorMsg    string
	callCount   int
}

func (m *e2eMockStockDataClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	m.callCount++

	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}

	price, ok := m.prices[stockCode]
	if !ok {
		price = 1000.0 // Default price
	}

	return &models.StockPrice{
		Code:       stockCode,
		Date:       time.Now(),
		OpenPrice:  client.FloatToDecimal(price * 0.98),
		HighPrice:  client.FloatToDecimal(price * 1.02),
		LowPrice:   client.FloatToDecimal(price * 0.97),
		ClosePrice: client.FloatToDecimal(price),
		Volume:     1000000,
		CreatedAt:  fixture.NullTimeFrom(time.Now()),
		UpdatedAt:  fixture.NullTimeFrom(time.Now()),
	}, nil
}

func (m *e2eMockStockDataClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return nil, nil
}

func (m *e2eMockStockDataClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return nil, nil
}

// Enhanced mock notification service
type e2eMockNotificationService struct {
	sendMessageCalled     bool
	sendDailyReportCalled bool
	sendStockAlertCalled  bool
	lastMessage           string
	lastTotalValue        float64
	lastTotalGain         float64
	lastGainPercent       float64
	shouldError           bool
	errorMsg              string
}

func (m *e2eMockNotificationService) SendMessage(message string) error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	m.sendMessageCalled = true
	m.lastMessage = message
	return nil
}

func (m *e2eMockNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	m.sendDailyReportCalled = true
	m.lastTotalValue = totalValue
	m.lastTotalGain = totalGain
	m.lastGainPercent = gainPercent
	return nil
}

func (m *e2eMockNotificationService) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	m.sendStockAlertCalled = true
	return nil
}
