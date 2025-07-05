package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/testutil/fixture"
)

func TestNewTechnicalAnalysisUseCase(t *testing.T) {
	stockRepo := NewMockStockRepository()
	stockClient := NewMockStockDataClient()

	uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)

	if uc == nil {
		t.Fatal("NewTechnicalAnalysisUseCase returned nil")
	}
}

func TestTechnicalAnalysisUseCase_CalculateAndSaveTechnicalIndicators(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		stockCode   string
		setupMocks  func(*MockStockRepository)
		expectError bool
	}{
		{
			name:      "Success with price history",
			stockCode: "7203",
			setupMocks: func(sr *MockStockRepository) {
				// Add price history
				for i := 0; i < 30; i++ {
					date := time.Now().AddDate(0, 0, -i)
					sr.stockPrices[date.Format("2006-01-02")] = &models.StockPrice{
						ID:         fixture.StockPriceID1,
						Code:       "7203",
						Date:       date,
						ClosePrice: client.FloatToDecimal(2000.0 + float64(i%5)*10),
						OpenPrice:  client.FloatToDecimal(1995.0),
						HighPrice:  client.FloatToDecimal(2010.0),
						LowPrice:   client.FloatToDecimal(1990.0),
						Volume:     1000000,
					}
				}
			},
			expectError: false,
		},
		{
			name:      "No price history",
			stockCode: "7203",
			setupMocks: func(sr *MockStockRepository) {
				// No price history
			},
			expectError: true, // Should error due to insufficient data
		},
		{
			name:      "Empty stock code",
			stockCode: "",
			setupMocks: func(sr *MockStockRepository) {
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := NewMockStockRepository()
			stockClient := NewMockStockDataClient()

			// Override GetPriceHistory for this test
			stockRepo.GetPriceHistoryFunc = func(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error) {
				var prices []*models.StockPrice
				for _, p := range stockRepo.stockPrices {
					if p.Code == stockCode {
						prices = append(prices, p)
					}
				}
				return prices, nil
			}

			tt.setupMocks(stockRepo)

			uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)
			err := uc.CalculateAndSaveTechnicalIndicators(ctx, tt.stockCode)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTechnicalAnalysisUseCase_GetTechnicalAnalysis(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	stockClient := NewMockStockDataClient()

	// Add test data
	stockRepo.technicalIndicators["7203"] = &models.TechnicalIndicator{
		ID:            "ind-1",
		Code:          "7203",
		Date:          time.Now(),
		Rsi14:         fixture.NullDecimalFrom(65.5),
		Macd:          fixture.NullDecimalFrom(10.5),
		MacdSignal:    fixture.NullDecimalFrom(8.5),
		MacdHistogram: fixture.NullDecimalFrom(2.0),
		Sma5:          fixture.NullDecimalFrom(2100.0),
		Sma25:         fixture.NullDecimalFrom(2050.0),
		Sma75:         fixture.NullDecimalFrom(2000.0),
	}

	uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)
	analysis, err := uc.GetTechnicalAnalysis(ctx, "7203")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected analysis but got nil")
	}

	if analysis.Code != "7203" {
		t.Errorf("Expected stock code 7203, got %s", analysis.Code)
	}
}

func TestTechnicalAnalysisUseCase_AnalyzeWatchList(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	stockClient := NewMockStockDataClient()

	// Setup watch list
	stockRepo.watchList = []*models.WatchList{
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

	// Add technical indicators
	stockRepo.technicalIndicators["7203"] = &models.TechnicalIndicator{
		Code:  "7203",
		Rsi14: fixture.NullDecimalFrom(30.0), // Oversold
	}
	stockRepo.technicalIndicators["6758"] = &models.TechnicalIndicator{
		Code:  "6758",
		Rsi14: fixture.NullDecimalFrom(70.0), // Overbought
	}

	uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)
	err := uc.AnalyzeWatchList(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// AnalyzeWatchList just processes stocks but doesn't return results
	// We can verify by checking if indicators were calculated
	// (In a real test, we might check logs or database state)
}

func TestTechnicalAnalysisUseCase_GetTradingSignals(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	stockClient := NewMockStockDataClient()

	// Setup watch list
	stockRepo.watchList = []*models.WatchList{
		{
			ID:       fixture.WatchListID1,
			Code:     "7203",
			Name:     "トヨタ自動車",
			IsActive: fixture.NullBoolFrom(true),
		},
	}

	// Add technical indicators with strong buy signal
	stockRepo.technicalIndicators["7203"] = &models.TechnicalIndicator{
		Code:          "7203",
		Rsi14:         fixture.NullDecimalFrom(25.0), // Very oversold
		Macd:          fixture.NullDecimalFrom(5.0),
		MacdSignal:    fixture.NullDecimalFrom(2.0),
		MacdHistogram: fixture.NullDecimalFrom(3.0), // Positive histogram
		Sma5:          fixture.NullDecimalFrom(2100.0),
		Sma25:         fixture.NullDecimalFrom(2000.0), // Uptrend
	}

	// Add current price
	stockRepo.stockPrices["7203"] = &models.StockPrice{
		Code:       "7203",
		Date:       time.Now(),
		ClosePrice: client.FloatToDecimal(2150.0),
	}

	uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)
	signals, err := uc.GetTradingSignals(ctx, "7203")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(signals) == 0 {
		t.Fatal("Expected at least one signal")
	}

	// Check for buy signals in Japanese
	buySignalFound := false
	for _, signal := range signals {
		if signal == "RSI買いシグナル（売られすぎ）" || signal == "MACDゴールデンクロス（買いシグナル）" {
			buySignalFound = true
			break
		}
	}

	if !buySignalFound {
		t.Error("Expected at least one buy signal")
	}
}

func TestTechnicalAnalysisUseCase_GetTradingSignals_NoIndicators(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	stockClient := NewMockStockDataClient()

	// Setup watch list but no indicators
	stockRepo.watchList = []*models.WatchList{
		{
			ID:       fixture.WatchListID1,
			Code:     "7203",
			Name:     "トヨタ自動車",
			IsActive: fixture.NullBoolFrom(true),
		},
	}

	uc := NewTechnicalAnalysisUseCase(stockRepo, stockClient)
	signals, err := uc.GetTradingSignals(ctx, "7203")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return empty signals when no indicators available
	if len(signals) != 0 {
		t.Errorf("Expected no signals, got %d", len(signals))
	}
}
