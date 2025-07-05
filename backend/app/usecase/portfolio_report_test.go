package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/analysis"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/google/go-cmp/cmp"
)

// Mock implementations for testing
type mockStockRepository struct {
	getLatestPriceFunc func(ctx context.Context, code string) (*models.StockPrice, error)
}

func (m *mockStockRepository) SaveStockPrice(ctx context.Context, price *models.StockPrice) error {
	return nil
}

func (m *mockStockRepository) SaveStockPrices(ctx context.Context, prices []*models.StockPrice) error {
	return nil
}

func (m *mockStockRepository) GetLatestPrice(ctx context.Context, stockCode string) (*models.StockPrice, error) {
	if m.getLatestPriceFunc != nil {
		return m.getLatestPriceFunc(ctx, stockCode)
	}
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) GetPriceHistory(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (m *mockStockRepository) CleanupOldData(ctx context.Context, days int) error {
	return nil
}

func (m *mockStockRepository) SaveTechnicalIndicator(ctx context.Context, indicator *models.TechnicalIndicator) error {
	return nil
}

func (m *mockStockRepository) GetLatestTechnicalIndicator(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error) {
	return nil, nil
}

func (m *mockStockRepository) GetActiveWatchList(ctx context.Context) ([]*models.WatchList, error) {
	return nil, nil
}

func (m *mockStockRepository) GetWatchListItem(ctx context.Context, id string) (*models.WatchList, error) {
	return nil, nil
}

func (m *mockStockRepository) AddToWatchList(ctx context.Context, item *models.WatchList) error {
	return nil
}

func (m *mockStockRepository) UpdateWatchList(ctx context.Context, item *models.WatchList) error {
	return nil
}

func (m *mockStockRepository) DeleteFromWatchList(ctx context.Context, id string) error {
	return nil
}

type mockPortfolioRepository struct {
	getAllFunc func(ctx context.Context) ([]*models.Portfolio, error)
}

func (m *mockPortfolioRepository) Create(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) GetByID(ctx context.Context, id string) (*models.Portfolio, error) {
	return nil, nil
}

func (m *mockPortfolioRepository) GetByCode(ctx context.Context, code string) (*models.Portfolio, error) {
	return nil, nil
}

func (m *mockPortfolioRepository) GetAll(ctx context.Context) ([]*models.Portfolio, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPortfolioRepository) Update(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockPortfolioRepository) GetTotalValue(ctx context.Context, currentPrices map[string]float64) (float64, error) {
	return 0, nil
}

func (m *mockPortfolioRepository) GetHoldingsByCode(ctx context.Context, codes []string) ([]*models.Portfolio, error) {
	return nil, nil
}

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

type mockNotificationService struct {
	sendMessageFunc     func(message string) error
	sendDailyReportFunc func(totalValue, totalGain, gainPercent float64) error
}

func (m *mockNotificationService) SendMessage(message string) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(message)
	}
	return nil
}

func (m *mockNotificationService) SendDailyReport(totalValue, totalGain, gainPercent float64) error {
	if m.sendDailyReportFunc != nil {
		return m.sendDailyReportFunc(totalValue, totalGain, gainPercent)
	}
	return nil
}

func (m *mockNotificationService) SendStockAlert(stockCode string, stockName string, currentPrice float64, changePercent float64, alertType string) error {
	return nil
}

func TestPortfolioReportUseCase_GenerateAndSendDailyReport(t *testing.T) {
	tests := []struct {
		name                string
		portfolio           []*models.Portfolio
		stockPrices         map[string]*models.StockPrice
		expectNotification  bool
		expectedTotalValue  float64
		expectedTotalGain   float64
		expectedGainPercent float64
		wantErr             bool
	}{
		{
			name: "successful report generation with profit",
			portfolio: []*models.Portfolio{
				{
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: client.FloatToDecimal(2200.0),
					Date:       time.Now(),
				},
			},
			expectNotification:  true,
			expectedTotalValue:  220000.0,
			expectedTotalGain:   20000.0,
			expectedGainPercent: 10.0,
			wantErr:             false,
		},
		{
			name: "successful report generation with loss",
			portfolio: []*models.Portfolio{
				{
					Code:          "9983",
					Name:          "ファーストリテイリング",
					Shares:        50,
					PurchasePrice: client.FloatToDecimal(50000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"9983": {
					Code:       "9983",
					ClosePrice: client.FloatToDecimal(48000.0),
					Date:       time.Now(),
				},
			},
			expectNotification:  true,
			expectedTotalValue:  2400000.0,
			expectedTotalGain:   -100000.0,
			expectedGainPercent: -4.0,
			wantErr:             false,
		},
		{
			name:               "empty portfolio",
			portfolio:          []*models.Portfolio{},
			stockPrices:        map[string]*models.StockPrice{},
			expectNotification: false,
			wantErr:            false,
		},
		{
			name: "portfolio with missing price data",
			portfolio: []*models.Portfolio{
				{
					Code:          "1111",
					Name:          "Missing Stock",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices:         map[string]*models.StockPrice{},
			expectNotification:  true,
			expectedTotalValue:  0.0,
			expectedTotalGain:   0.0,
			expectedGainPercent: 0.0,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var notificationCalled bool
			var sentTotalValue, sentTotalGain, sentGainPercent float64

			stockRepo := &mockStockRepository{
				getLatestPriceFunc: func(ctx context.Context, code string) (*models.StockPrice, error) {
					if price, ok := tt.stockPrices[code]; ok {
						return price, nil
					}
					return nil, errors.New("price not found")
				},
			}

			portfolioRepo := &mockPortfolioRepository{
				getAllFunc: func(ctx context.Context) ([]*models.Portfolio, error) {
					return tt.portfolio, nil
				},
			}

			notifier := &mockNotificationService{
				sendDailyReportFunc: func(totalValue, totalGain, gainPercent float64) error {
					notificationCalled = true
					sentTotalValue = totalValue
					sentTotalGain = totalGain
					sentGainPercent = gainPercent
					return nil
				},
			}

			uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, notifier)

			err := uc.GenerateAndSendDailyReport(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAndSendDailyReport() error = %v, wantErr %v", err, tt.wantErr)
			}

			if notificationCalled != tt.expectNotification {
				t.Errorf("notification called = %v, expected %v", notificationCalled, tt.expectNotification)
			}

			if tt.expectNotification && notificationCalled {
				const tolerance = 0.01
				if diff := sentTotalValue - tt.expectedTotalValue; diff > tolerance || diff < -tolerance {
					t.Errorf("total value = %v, expected %v", sentTotalValue, tt.expectedTotalValue)
				}
				if diff := sentTotalGain - tt.expectedTotalGain; diff > tolerance || diff < -tolerance {
					t.Errorf("total gain = %v, expected %v", sentTotalGain, tt.expectedTotalGain)
				}
				if diff := sentGainPercent - tt.expectedGainPercent; diff > tolerance || diff < -tolerance {
					t.Errorf("gain percent = %v, expected %v", sentGainPercent, tt.expectedGainPercent)
				}
			}
		})
	}
}

func TestPortfolioReportUseCase_GenerateComprehensiveDailyReport(t *testing.T) {
	tests := []struct {
		name             string
		portfolio        []*models.Portfolio
		stockPrices      map[string]*models.StockPrice
		expectedContains []string
		wantErr          bool
	}{
		{
			name: "comprehensive report with mixed performance",
			portfolio: []*models.Portfolio{
				{
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
				{
					Code:          "9983",
					Name:          "ファーストリテイリング",
					Shares:        50,
					PurchasePrice: client.FloatToDecimal(50000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: client.FloatToDecimal(2200.0),
					Date:       time.Now(),
				},
				"9983": {
					Code:       "9983",
					ClosePrice: client.FloatToDecimal(48000.0),
					Date:       time.Now(),
				},
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
			portfolio: []*models.Portfolio{},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
				"💡 現在ポートフォリオにデータがありません",
			},
			wantErr: false,
		},
		{
			name: "report with price errors",
			portfolio: []*models.Portfolio{
				{
					Code:          "1111",
					Name:          "エラー銘柄",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{},
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
			stockRepo := &mockStockRepository{
				getLatestPriceFunc: func(ctx context.Context, code string) (*models.StockPrice, error) {
					if price, ok := tt.stockPrices[code]; ok {
						return price, nil
					}
					return nil, errors.New("price not found")
				},
			}

			portfolioRepo := &mockPortfolioRepository{
				getAllFunc: func(ctx context.Context) ([]*models.Portfolio, error) {
					return tt.portfolio, nil
				},
			}

			uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, &mockNotificationService{})

			report, err := uc.GenerateComprehensiveDailyReport(context.Background())

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
	tests := []struct {
		name            string
		portfolio       []*models.Portfolio
		stockPrices     map[string]*models.StockPrice
		expectedSummary *analysis.PortfolioSummary
		wantErr         bool
	}{
		{
			name: "statistics with valid data",
			portfolio: []*models.Portfolio{
				{
					Code:          "7203",
					Name:          "トヨタ自動車",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: client.FloatToDecimal(2200.0),
					Date:       time.Now(),
				},
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
			portfolio: []*models.Portfolio{},
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
			stockRepo := &mockStockRepository{
				getLatestPriceFunc: func(ctx context.Context, code string) (*models.StockPrice, error) {
					if price, ok := tt.stockPrices[code]; ok {
						return price, nil
					}
					return nil, errors.New("price not found")
				},
			}

			portfolioRepo := &mockPortfolioRepository{
				getAllFunc: func(ctx context.Context) ([]*models.Portfolio, error) {
					return tt.portfolio, nil
				},
			}

			uc := NewPortfolioReportUseCase(stockRepo, portfolioRepo, &mockStockDataClient{}, &mockNotificationService{})

			summary, err := uc.GetPortfolioStatistics(context.Background())

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
