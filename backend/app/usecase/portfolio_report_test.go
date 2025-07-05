package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/testutil/fixture"
)

// MockNotificationService for testing
type MockNotificationService struct {
	sendMessageCalled     bool
	sendDailyReportCalled bool
	sendStockAlertCalled  bool
	lastMessage           string
	lastTotalValue        float64
	lastTotalGain         float64
	lastGainPercent       float64
	sendMessageErr        error
	sendDailyReportErr    error
}

func (m *MockNotificationService) SendMessage(message string) error {
	m.sendMessageCalled = true
	m.lastMessage = message
	return m.sendMessageErr
}

func (m *MockNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	m.sendDailyReportCalled = true
	m.lastTotalValue = totalValue
	m.lastTotalGain = totalGain
	m.lastGainPercent = gainPercent
	return m.sendDailyReportErr
}

func (m *MockNotificationService) SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
	m.sendStockAlertCalled = true
	return nil
}

func TestNewPortfolioReportUseCase(t *testing.T) {
	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()
	notificationService := &MockNotificationService{}

	uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notificationService)

	if uc == nil {
		t.Fatal("NewPortfolioReportUseCase returned nil")
	}
}

func TestPortfolioReportUseCase_GenerateAndSendDailyReport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMocks  func(*MockStockRepository, *MockPortfolioRepository, *MockStockDataClient, *MockNotificationService)
		expectError bool
	}{
		{
			name: "Success with portfolio",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository, sc *MockStockDataClient, ns *MockNotificationService) {
				// Setup portfolio
				pr.portfolios["p1"] = &models.Portfolio{
					ID:            "p1",
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now().AddDate(0, -6, 0),
				}

				// Setup current price
				sc.prices["7203"] = 2150.0

				// Also setup stock repo to return price
				sr.stockPrices["7203"] = &models.StockPrice{
					Code:       "7203",
					ClosePrice: client.FloatToDecimal(2150.0),
					Date:       time.Now(),
				}
			},
			expectError: false,
		},
		{
			name: "Empty portfolio",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository, sc *MockStockDataClient, ns *MockNotificationService) {
				// No portfolio
			},
			expectError: false,
		},
		{
			name: "Notification error",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository, sc *MockStockDataClient, ns *MockNotificationService) {
				pr.portfolios["p1"] = &models.Portfolio{
					ID:            "p1",
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now().AddDate(0, -6, 0),
				}
				sc.prices["7203"] = 2150.0
				sr.stockPrices["7203"] = &models.StockPrice{
					Code:       "7203",
					ClosePrice: client.FloatToDecimal(2150.0),
					Date:       time.Now(),
				}
				ns.sendDailyReportErr = fmt.Errorf("notification failed")
			},
			expectError: true,
		},
		{
			name: "Stock price fetch error",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository, sc *MockStockDataClient, ns *MockNotificationService) {
				pr.portfolios["p1"] = &models.Portfolio{
					ID:            "p1",
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now().AddDate(0, -6, 0),
				}
				sc.getPriceErr = fmt.Errorf("API error")
			},
			expectError: false, // Should continue with 0 price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := NewMockStockRepository()
			portfolioRepo := NewMockPortfolioRepository()
			stockClient := NewMockStockDataClient()
			notificationService := &MockNotificationService{}

			tt.setupMocks(stockRepo, portfolioRepo, stockClient, notificationService)

			uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notificationService)
			err := uc.GenerateAndSendDailyReport(ctx)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && len(portfolioRepo.portfolios) > 0 {
				if !notificationService.sendDailyReportCalled {
					t.Error("Daily report was not sent")
				}

				// Verify calculations
				if notificationService.lastTotalValue > 0 {
					expectedValue := 2150.0 * 100 // price * shares
					if notificationService.lastTotalValue != expectedValue {
						t.Errorf("Expected total value %f, got %f", expectedValue, notificationService.lastTotalValue)
					}

					expectedGain := (2150.0 - 2000.0) * 100 // (current - purchase) * shares
					if notificationService.lastTotalGain != expectedGain {
						t.Errorf("Expected total gain %f, got %f", expectedGain, notificationService.lastTotalGain)
					}
				}
			}
		})
	}
}

func TestPortfolioReportUseCase_SendComprehensiveDailyReport(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()
	notificationService := &MockNotificationService{}

	// Setup test data
	portfolioRepo.portfolios["p1"] = &models.Portfolio{
		ID:            "p1",
		Code:          "7203",
		Name:          "トヨタ自動車",
		Shares:        100,
		PurchasePrice: client.FloatToDecimal(2000.0),
		PurchaseDate:  time.Now().AddDate(0, -6, 0),
	}
	portfolioRepo.portfolios["p2"] = &models.Portfolio{
		ID:            "p2",
		Code:          "6758",
		Name:          "ソニーグループ",
		Shares:        50,
		PurchasePrice: client.FloatToDecimal(12000.0),
		PurchaseDate:  time.Now().AddDate(0, -3, 0),
	}

	stockClient.prices["7203"] = 2150.0
	stockClient.prices["6758"] = 13500.0

	// Setup stock repo to return prices
	stockRepo.stockPrices["7203"] = &models.StockPrice{
		Code:       "7203",
		ClosePrice: client.FloatToDecimal(2150.0),
		Date:       time.Now(),
	}
	stockRepo.stockPrices["6758"] = &models.StockPrice{
		Code:       "6758",
		ClosePrice: client.FloatToDecimal(13500.0),
		Date:       time.Now(),
	}

	// Add technical indicators
	stockRepo.technicalIndicators["7203"] = &models.TechnicalIndicator{
		Code:  "7203",
		Rsi14: fixture.NullDecimalFrom(65.0),
		Sma5:  fixture.NullDecimalFrom(2100.0),
		Sma25: fixture.NullDecimalFrom(2050.0),
	}

	uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notificationService)
	err := uc.SendComprehensiveDailyReport(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !notificationService.sendMessageCalled {
		t.Error("Comprehensive report was not sent")
	}

	// Verify report contains expected content
	report := notificationService.lastMessage
	if report == "" {
		t.Error("Report is empty")
	}

	// Check for key elements
	expectedElements := []string{
		"📊 ポートフォリオレポート",
		"トヨタ自動車",
		"ソニーグループ",
		"現在価値:",
		"損益:",
	}

	for _, element := range expectedElements {
		if !contains(report, element) {
			t.Errorf("Report missing expected element: %s", element)
		}
	}
}

func TestPortfolioReportUseCase_GetPortfolioStatistics(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()
	notificationService := &MockNotificationService{}

	// Setup portfolio with different performance
	portfolioRepo.portfolios["p1"] = &models.Portfolio{
		ID:            "p1",
		Code:          "7203",
		Name:          "トヨタ自動車",
		Shares:        100,
		PurchasePrice: client.FloatToDecimal(2000.0),
		PurchaseDate:  time.Now().AddDate(0, -6, 0),
	}
	portfolioRepo.portfolios["p2"] = &models.Portfolio{
		ID:            "p2",
		Code:          "6758",
		Name:          "ソニーグループ",
		Shares:        50,
		PurchasePrice: client.FloatToDecimal(13000.0), // Loss position
		PurchaseDate:  time.Now().AddDate(0, -3, 0),
	}

	stockClient.prices["7203"] = 2300.0  // 15% gain
	stockClient.prices["6758"] = 12000.0 // -7.7% loss

	// Setup stock repo to return prices
	stockRepo.stockPrices["7203"] = &models.StockPrice{
		Code:       "7203",
		ClosePrice: client.FloatToDecimal(2300.0),
		Date:       time.Now(),
	}
	stockRepo.stockPrices["6758"] = &models.StockPrice{
		Code:       "6758",
		ClosePrice: client.FloatToDecimal(12000.0),
		Date:       time.Now(),
	}

	uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notificationService)
	stats, err := uc.GetPortfolioStatistics(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected statistics but got nil")
	}

	// Verify statistics
	if len(stats.Holdings) != 2 {
		t.Errorf("Expected 2 holdings, got %d", len(stats.Holdings))
	}

	expectedTotalValue := 2300.0*100 + 12000.0*50
	if stats.TotalValue != expectedTotalValue {
		t.Errorf("Expected total value %f, got %f", expectedTotalValue, stats.TotalValue)
	}

	// Find top and worst performer
	var topCode, worstCode string
	var topGain, worstGain float64 = -999999, 999999

	for _, holding := range stats.Holdings {
		if holding.GainPercent > topGain {
			topGain = holding.GainPercent
			topCode = holding.Code
		}
		if holding.GainPercent < worstGain {
			worstGain = holding.GainPercent
			worstCode = holding.Code
		}
	}

	if topCode != "7203" {
		t.Errorf("Top performer should be 7203, got %s", topCode)
	}

	if worstCode != "6758" {
		t.Errorf("Worst performer should be 6758, got %s", worstCode)
	}
}

func TestPortfolioReportUseCase_HandlePriceErrors(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()
	notificationService := &MockNotificationService{}

	// Setup portfolio
	portfolioRepo.portfolios["p1"] = &models.Portfolio{
		ID:            "p1",
		Code:          "7203",
		Name:          "トヨタ自動車",
		Shares:        100,
		PurchasePrice: client.FloatToDecimal(2000.0),
		PurchaseDate:  time.Now().AddDate(0, -6, 0),
	}

	// Simulate price fetch error for specific stock
	stockClient.getPriceErr = fmt.Errorf("API rate limit")

	uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, stockClient, notificationService)
	err := uc.GenerateAndSendDailyReport(ctx)

	if err != nil {
		t.Errorf("Should handle price errors gracefully: %v", err)
	}

	// Report should still be sent with available data
	if !notificationService.sendDailyReportCalled {
		t.Error("Report should be sent even with price errors")
	}

	// Value should be 0 for stocks with price errors
	if notificationService.lastTotalValue != 0 {
		t.Errorf("Expected total value 0 when price fetch fails, got %f", notificationService.lastTotalValue)
	}
}
