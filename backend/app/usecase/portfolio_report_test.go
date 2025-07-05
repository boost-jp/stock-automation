package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/analysis"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing
type mockStockRepository struct {
	getLatestPriceFunc func(ctx context.Context, code string) (*models.StockPrice, error)
}

func (m *mockStockRepository) BulkInsertStockPrices(ctx context.Context, prices []*models.StockPrice) error {
	return nil
}

func (m *mockStockRepository) GetAll(ctx context.Context) ([]*models.WatchList, error) {
	return nil, nil
}

func (m *mockStockRepository) AddWatchItem(ctx context.Context, code, name string) error {
	return nil
}

func (m *mockStockRepository) RemoveWatchItem(ctx context.Context, code string) error {
	return nil
}

func (m *mockStockRepository) GetLatestPrice(ctx context.Context, code string) (*models.StockPrice, error) {
	if m.getLatestPriceFunc != nil {
		return m.getLatestPriceFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) SaveTechnicalIndicators(ctx context.Context, indicators []*models.TechnicalIndicator) error {
	return nil
}

func (m *mockStockRepository) GetTechnicalIndicator(ctx context.Context, code string, indicatorType string, date time.Time) (*models.TechnicalIndicator, error) {
	return nil, nil
}

func (m *mockStockRepository) GetHistoricalPrices(ctx context.Context, code string, startDate, endDate time.Time) ([]*models.StockPrice, error) {
	return nil, nil
}

type mockPortfolioRepository struct {
	getAllFunc func(ctx context.Context) ([]*models.Portfolio, error)
}

func (m *mockPortfolioRepository) GetAll(ctx context.Context) ([]*models.Portfolio, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPortfolioRepository) Add(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) Remove(ctx context.Context, code string) error {
	return nil
}

func (m *mockPortfolioRepository) Update(ctx context.Context, code string, shares int, purchasePrice float64) error {
	return nil
}

type mockStockDataClient struct{}

func (m *mockStockDataClient) GetStockPrice(code string) (*client.StockData, error) {
	return nil, nil
}

func (m *mockStockDataClient) GetBulkPrices(codes []string) (map[string]*client.StockData, error) {
	return nil, nil
}

func (m *mockStockDataClient) GetHistoricalData(code string, startDate, endDate time.Time) ([]*client.HistoricalData, error) {
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(2000.0)),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: types.NewDecimal(decimal.NewFromFloat(2200.0)),
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(50000.0)),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"9983": {
					Code:       "9983",
					ClosePrice: types.NewDecimal(decimal.NewFromFloat(48000.0)),
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(1000.0)),
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(2000.0)),
					PurchaseDate:  time.Now(),
				},
				{
					Code:          "9983",
					Name:          "ファーストリテイリング",
					Shares:        50,
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(50000.0)),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: types.NewDecimal(decimal.NewFromFloat(2200.0)),
					Date:       time.Now(),
				},
				"9983": {
					Code:       "9983",
					ClosePrice: types.NewDecimal(decimal.NewFromFloat(48000.0)),
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(1000.0)),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
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
						t.Errorf("report does not contain expected string: %s", expected)
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
					PurchasePrice: types.NewDecimal(decimal.NewFromFloat(2000.0)),
					PurchaseDate:  time.Now(),
				},
			},
			stockPrices: map[string]*models.StockPrice{
				"7203": {
					Code:       "7203",
					ClosePrice: types.NewDecimal(decimal.NewFromFloat(2200.0)),
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
