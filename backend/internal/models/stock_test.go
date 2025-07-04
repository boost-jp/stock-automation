package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStockPrice_Validation(t *testing.T) {
	tests := []struct {
		name        string
		stockPrice  StockPrice
		expectValid bool
	}{
		{
			name: "Valid stock price",
			stockPrice: StockPrice{
				Code:      "1234",
				Name:      "Test Stock",
				Price:     1000.0,
				Volume:    100000,
				High:      1050.0,
				Low:       950.0,
				Open:      980.0,
				Close:     1000.0,
				Timestamp: time.Now(),
			},
			expectValid: true,
		},
		{
			name: "Empty code",
			stockPrice: StockPrice{
				Code:      "",
				Name:      "Test Stock",
				Price:     1000.0,
				Volume:    100000,
				High:      1050.0,
				Low:       950.0,
				Open:      980.0,
				Close:     1000.0,
				Timestamp: time.Now(),
			},
			expectValid: false,
		},
		{
			name: "Negative price",
			stockPrice: StockPrice{
				Code:      "1234",
				Name:      "Test Stock",
				Price:     -1000.0,
				Volume:    100000,
				High:      1050.0,
				Low:       950.0,
				Open:      980.0,
				Close:     1000.0,
				Timestamp: time.Now(),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.stockPrice.IsValid()
			assert.Equal(t, tt.expectValid, valid, "Stock price validation failed")
		})
	}
}

func TestTechnicalIndicator_Validation(t *testing.T) {
	tests := []struct {
		name        string
		indicator   TechnicalIndicator
		expectValid bool
	}{
		{
			name: "Valid technical indicator",
			indicator: TechnicalIndicator{
				Code:      "1234",
				MA5:       1000.0,
				MA25:      1050.0,
				MA75:      1100.0,
				RSI:       65.5,
				MACD:      12.5,
				Signal:    10.0,
				Histogram: 2.5,
				Timestamp: time.Now(),
			},
			expectValid: true,
		},
		{
			name: "Invalid RSI range",
			indicator: TechnicalIndicator{
				Code:      "1234",
				MA5:       1000.0,
				MA25:      1050.0,
				MA75:      1100.0,
				RSI:       105.0, // RSI should be 0-100
				MACD:      12.5,
				Signal:    10.0,
				Histogram: 2.5,
				Timestamp: time.Now(),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.indicator.IsValid()
			assert.Equal(t, tt.expectValid, valid, "Technical indicator validation failed")
		})
	}
}

func TestPortfolio_Validation(t *testing.T) {
	tests := []struct {
		name        string
		portfolio   Portfolio
		expectValid bool
	}{
		{
			name: "Valid portfolio",
			portfolio: Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        100,
				PurchasePrice: 1000.0,
				PurchaseDate:  time.Now(),
			},
			expectValid: true,
		},
		{
			name: "Negative shares",
			portfolio: Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        -100,
				PurchasePrice: 1000.0,
				PurchaseDate:  time.Now(),
			},
			expectValid: false,
		},
		{
			name: "Zero purchase price",
			portfolio: Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        100,
				PurchasePrice: 0.0,
				PurchaseDate:  time.Now(),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.portfolio.IsValid()
			assert.Equal(t, tt.expectValid, valid, "Portfolio validation failed")
		})
	}
}

func TestWatchList_Validation(t *testing.T) {
	tests := []struct {
		name        string
		watchList   WatchList
		expectValid bool
	}{
		{
			name: "Valid watch list",
			watchList: WatchList{
				Code:            "1234",
				Name:            "Test Stock",
				TargetBuyPrice:  900.0,
				TargetSellPrice: 1100.0,
				IsActive:        true,
			},
			expectValid: true,
		},
		{
			name: "Invalid price range",
			watchList: WatchList{
				Code:            "1234",
				Name:            "Test Stock",
				TargetBuyPrice:  1100.0, // Buy price higher than sell price
				TargetSellPrice: 900.0,
				IsActive:        true,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.watchList.IsValid()
			assert.Equal(t, tt.expectValid, valid, "Watch list validation failed")
		})
	}
}

func TestStockPrice_CalculateGainLoss(t *testing.T) {
	tests := []struct {
		name           string
		currentPrice   float64
		purchasePrice  float64
		expectedGain   float64
		expectedReturn float64
	}{
		{
			name:           "Positive gain",
			currentPrice:   1100.0,
			purchasePrice:  1000.0,
			expectedGain:   100.0,
			expectedReturn: 10.0,
		},
		{
			name:           "Negative gain (loss)",
			currentPrice:   900.0,
			purchasePrice:  1000.0,
			expectedGain:   -100.0,
			expectedReturn: -10.0,
		},
		{
			name:           "No change",
			currentPrice:   1000.0,
			purchasePrice:  1000.0,
			expectedGain:   0.0,
			expectedReturn: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockPrice := StockPrice{
				Price: tt.currentPrice,
			}
			portfolio := Portfolio{
				PurchasePrice: tt.purchasePrice,
			}

			gain := stockPrice.CalculateGainLoss(portfolio)
			returnRate := stockPrice.CalculateReturnRate(portfolio)

			assert.Equal(t, tt.expectedGain, gain, "Gain calculation failed")
			assert.Equal(t, tt.expectedReturn, returnRate, "Return rate calculation failed")
		})
	}
}

func TestTechnicalIndicator_SignalStrength(t *testing.T) {
	tests := []struct {
		name             string
		indicator        TechnicalIndicator
		expectedStrength string
	}{
		{
			name: "Strong buy signal",
			indicator: TechnicalIndicator{
				RSI:       25.0, // Oversold
				MACD:      5.0,
				Signal:    2.0,
				Histogram: 3.0, // Positive histogram
			},
			expectedStrength: "Strong Buy",
		},
		{
			name: "Strong sell signal",
			indicator: TechnicalIndicator{
				RSI:       75.0, // Overbought
				MACD:      -5.0,
				Signal:    -2.0,
				Histogram: -3.0, // Negative histogram
			},
			expectedStrength: "Strong Sell",
		},
		{
			name: "Neutral signal",
			indicator: TechnicalIndicator{
				RSI:       50.0, // Neutral
				MACD:      1.0,
				Signal:    0.5,
				Histogram: 0.5,
			},
			expectedStrength: "Neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := tt.indicator.GetSignalStrength()
			assert.Equal(t, tt.expectedStrength, strength, "Signal strength calculation failed")
		})
	}
}

func TestPortfolio_TotalValue(t *testing.T) {
	tests := []struct {
		name          string
		portfolio     Portfolio
		currentPrice  float64
		expectedValue float64
	}{
		{
			name: "Calculate total value",
			portfolio: Portfolio{
				Shares:        100,
				PurchasePrice: 1000.0,
			},
			currentPrice:  1100.0,
			expectedValue: 110000.0, // 100 shares * 1100 price
		},
		{
			name: "Calculate total value with loss",
			portfolio: Portfolio{
				Shares:        50,
				PurchasePrice: 2000.0,
			},
			currentPrice:  1800.0,
			expectedValue: 90000.0, // 50 shares * 1800 price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalValue := tt.portfolio.CalculateTotalValue(tt.currentPrice)
			assert.Equal(t, tt.expectedValue, totalValue, "Total value calculation failed")
		})
	}
}

func TestPortfolio_PurchaseValue(t *testing.T) {
	tests := []struct {
		name                  string
		portfolio             Portfolio
		expectedPurchaseValue float64
	}{
		{
			name: "Calculate purchase value",
			portfolio: Portfolio{
				Shares:        100,
				PurchasePrice: 1000.0,
			},
			expectedPurchaseValue: 100000.0, // 100 shares * 1000 price
		},
		{
			name: "Calculate purchase value with fractional shares",
			portfolio: Portfolio{
				Shares:        50,
				PurchasePrice: 1500.5,
			},
			expectedPurchaseValue: 75025.0, // 50 shares * 1500.5 price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			purchaseValue := tt.portfolio.GetPurchaseValue()
			assert.Equal(t, tt.expectedPurchaseValue, purchaseValue, "Purchase value calculation failed")
		})
	}
}
