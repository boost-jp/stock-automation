package integration

import (
	"context"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/analysis"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/testutil"
	"github.com/boost-jp/stock-automation/app/usecase"
	"github.com/google/go-cmp/cmp"
	"github.com/oklog/ulid/v2"
)

// mockStockDataClient implements client.StockDataClient for testing (external API mock)
type mockStockDataClient struct{}

func (m *mockStockDataClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	return nil, nil
}

func (m *mockStockDataClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (m *mockStockDataClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

// mockNotificationService implements notification.NotificationService for testing (external service mock)
type mockNotificationService struct {
	sendMessageCalled     bool
	sendDailyReportCalled bool
	lastMessage           string
	lastTotalValue        float64
	lastTotalGain         float64
	lastGainPercent       float64
}

func (m *mockNotificationService) SendMessage(message string) error {
	m.sendMessageCalled = true
	m.lastMessage = message
	return nil
}

func (m *mockNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	m.sendDailyReportCalled = true
	m.lastTotalValue = totalValue
	m.lastTotalGain = totalGain
	m.lastGainPercent = gainPercent
	return nil
}

func (m *mockNotificationService) SendStockAlert(stockCode string, stockName string, currentPrice float64, changePercent float64, alertType string) error {
	return nil
}

func TestPortfolioReportUseCase_GenerateAndSendDailyReport(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	testDB := testutil.NewTestDB(t)
	defer testDB.Cleanup()

	// Initialize repositories with real database
	stockRepo := repository.NewStockRepository(testDB.GetBoilDB())
	portfolioRepo := repository.NewPortfolioRepository(testDB.GetBoilDB())

	tests := []struct {
		name                string
		setupFunc           func(t *testing.T)
		expectedTotalValue  float64
		expectedTotalGain   float64
		expectedGainPercent float64
		expectNotification  bool
		wantErr             bool
	}{
		{
			name: "successful report generation with profit",
			setupFunc: func(t *testing.T) {
				// Insert test portfolio
				portfolioID := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID, "7203", "トヨタ自動車", 100, 2000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio: %v", err)
				}

				// Insert current stock price
				err = testDB.InsertTestStockPrice(ctx, "7203", time.Now(), 2200.0, 2250.0, 2180.0, 2200.0, 1000000)
				if err != nil {
					t.Fatalf("Failed to insert test stock price: %v", err)
				}
			},
			expectedTotalValue:  220000.0,
			expectedTotalGain:   20000.0,
			expectedGainPercent: 10.0,
			expectNotification:  true,
			wantErr:             false,
		},
		{
			name: "successful report generation with loss",
			setupFunc: func(t *testing.T) {
				// Insert test portfolio
				portfolioID := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID, "9983", "ファーストリテイリング", 50, 50000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio: %v", err)
				}

				// Insert current stock price
				err = testDB.InsertTestStockPrice(ctx, "9983", time.Now(), 48000.0, 48500.0, 47500.0, 48000.0, 500000)
				if err != nil {
					t.Fatalf("Failed to insert test stock price: %v", err)
				}
			},
			expectedTotalValue:  2400000.0,
			expectedTotalGain:   -100000.0,
			expectedGainPercent: -4.0,
			expectNotification:  true,
			wantErr:             false,
		},
		{
			name:               "empty portfolio",
			setupFunc:          func(t *testing.T) {},
			expectNotification: false,
			wantErr:            false,
		},
		{
			name: "portfolio with missing price data",
			setupFunc: func(t *testing.T) {
				// Insert test portfolio without corresponding price data
				portfolioID := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID, "1111", "Missing Stock", 100, 1000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio: %v", err)
				}
			},
			expectedTotalValue:  0.0,
			expectedTotalGain:   0.0,
			expectedGainPercent: 0.0,
			expectNotification:  true,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up database before each test
			if err := testDB.TruncateAll(); err != nil {
				t.Fatalf("Failed to truncate tables: %v", err)
			}

			// Setup test data
			tt.setupFunc(t)

			// Create mock notification service
			notifier := &mockNotificationService{}

			// Create use case with real repositories and mock external services
			uc := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, notifier)

			// Execute test
			err := uc.GenerateAndSendDailyReport(ctx)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAndSendDailyReport() error = %v, wantErr %v", err, tt.wantErr)
			}

			if notifier.sendDailyReportCalled != tt.expectNotification {
				t.Errorf("notification called = %v, expected %v", notifier.sendDailyReportCalled, tt.expectNotification)
			}

			if tt.expectNotification && notifier.sendDailyReportCalled {
				const tolerance = 0.01
				if diff := notifier.lastTotalValue - tt.expectedTotalValue; diff > tolerance || diff < -tolerance {
					t.Errorf("total value = %v, expected %v", notifier.lastTotalValue, tt.expectedTotalValue)
				}
				if diff := notifier.lastTotalGain - tt.expectedTotalGain; diff > tolerance || diff < -tolerance {
					t.Errorf("total gain = %v, expected %v", notifier.lastTotalGain, tt.expectedTotalGain)
				}
				if diff := notifier.lastGainPercent - tt.expectedGainPercent; diff > tolerance || diff < -tolerance {
					t.Errorf("gain percent = %v, expected %v", notifier.lastGainPercent, tt.expectedGainPercent)
				}
			}
		})
	}
}

func TestPortfolioReportUseCase_GenerateComprehensiveDailyReport(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	testDB := testutil.NewTestDB(t)
	defer testDB.Cleanup()

	// Initialize repositories with real database
	stockRepo := repository.NewStockRepository(testDB.GetBoilDB())
	portfolioRepo := repository.NewPortfolioRepository(testDB.GetBoilDB())

	tests := []struct {
		name             string
		setupFunc        func(t *testing.T)
		expectedContains []string
		wantErr          bool
	}{
		{
			name: "comprehensive report with mixed performance",
			setupFunc: func(t *testing.T) {
				// Insert multiple portfolios
				portfolioID1 := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID1, "7203", "トヨタ自動車", 100, 2000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio 1: %v", err)
				}

				portfolioID2 := ulid.MustNew(ulid.Now(), nil).String()
				err = testDB.InsertTestPortfolio(ctx, portfolioID2, "9983", "ファーストリテイリング", 50, 50000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio 2: %v", err)
				}

				// Insert stock prices
				err = testDB.InsertTestStockPrice(ctx, "7203", time.Now(), 2200.0, 2250.0, 2180.0, 2200.0, 1000000)
				if err != nil {
					t.Fatalf("Failed to insert stock price 1: %v", err)
				}

				err = testDB.InsertTestStockPrice(ctx, "9983", time.Now(), 48000.0, 48500.0, 47500.0, 48000.0, 500000)
				if err != nil {
					t.Fatalf("Failed to insert stock price 2: %v", err)
				}
			},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
				"💰 総資産状況",
				"トヨタ自動車",
				"ファーストリテイリング",
				"🕐 生成時刻:",
			},
			wantErr: false,
		},
		{
			name:      "empty portfolio report",
			setupFunc: func(t *testing.T) {},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
				"💡 現在ポートフォリオにデータがありません",
			},
			wantErr: false,
		},
		{
			name: "report with price errors",
			setupFunc: func(t *testing.T) {
				// Insert portfolio without price data
				portfolioID := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID, "1111", "エラー銘柄", 100, 1000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio: %v", err)
				}
			},
			expectedContains: []string{
				"ポートフォリオにデータがありません",
				"⚠️ 価格取得エラー:",
				"エラー銘柄 (1111): 価格取得エラー",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up database before each test
			if err := testDB.TruncateAll(); err != nil {
				t.Fatalf("Failed to truncate tables: %v", err)
			}

			// Setup test data
			tt.setupFunc(t)

			// Create use case with real repositories and mock external services
			uc := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, &mockNotificationService{})

			// Execute test
			report, err := uc.GenerateComprehensiveDailyReport(ctx)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateComprehensiveDailyReport() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				for _, expected := range tt.expectedContains {
					if !contains(report, expected) {
						t.Errorf("report does not contain expected string: %s\nActual report:\n%s", expected, report)
					}
				}
			}
		})
	}
}

func TestPortfolioReportUseCase_GetPortfolioStatistics(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	testDB := testutil.NewTestDB(t)
	defer testDB.Cleanup()

	// Initialize repositories with real database
	stockRepo := repository.NewStockRepository(testDB.GetBoilDB())
	portfolioRepo := repository.NewPortfolioRepository(testDB.GetBoilDB())

	tests := []struct {
		name            string
		setupFunc       func(t *testing.T)
		expectedSummary *analysis.PortfolioSummary
		wantErr         bool
	}{
		{
			name: "statistics with valid data",
			setupFunc: func(t *testing.T) {
				// Insert test portfolio
				portfolioID := ulid.MustNew(ulid.Now(), nil).String()
				err := testDB.InsertTestPortfolio(ctx, portfolioID, "7203", "トヨタ自動車", 100, 2000.0, time.Now())
				if err != nil {
					t.Fatalf("Failed to insert test portfolio: %v", err)
				}

				// Insert current stock price
				err = testDB.InsertTestStockPrice(ctx, "7203", time.Now(), 2200.0, 2250.0, 2180.0, 2200.0, 1000000)
				if err != nil {
					t.Fatalf("Failed to insert test stock price: %v", err)
				}
			},
			expectedSummary: &analysis.PortfolioSummary{
				TotalValue:       220000.0,
				TotalCost:        200000.0,
				TotalGain:        20000.0,
				TotalGainPercent: 10.0,
			},
			wantErr: false,
		},
		{
			name:      "empty portfolio statistics",
			setupFunc: func(t *testing.T) {},
			expectedSummary: &analysis.PortfolioSummary{
				TotalValue:       0.0,
				TotalCost:        0.0,
				TotalGain:        0.0,
				TotalGainPercent: 0.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up database before each test
			if err := testDB.TruncateAll(); err != nil {
				t.Fatalf("Failed to truncate tables: %v", err)
			}

			// Setup test data
			tt.setupFunc(t)

			// Create use case with real repositories and mock external services
			uc := usecase.NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, &mockNotificationService{})

			// Execute test
			summary, err := uc.GetPortfolioStatistics(ctx)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPortfolioStatistics() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && summary != nil {
				// Compare key fields
				if diff := cmp.Diff(tt.expectedSummary.TotalValue, summary.TotalValue); diff != "" {
					t.Errorf("TotalValue mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expectedSummary.TotalCost, summary.TotalCost); diff != "" {
					t.Errorf("TotalCost mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expectedSummary.TotalGain, summary.TotalGain); diff != "" {
					t.Errorf("TotalGain mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expectedSummary.TotalGainPercent, summary.TotalGainPercent); diff != "" {
					t.Errorf("TotalGainPercent mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
