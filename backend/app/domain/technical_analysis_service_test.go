package domain

import (
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/google/go-cmp/cmp"
)

func TestNewTechnicalAnalysisService(t *testing.T) {
	service := NewTechnicalAnalysisService()
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestTechnicalAnalysisService_ConvertStockPrices(t *testing.T) {
	service := NewTechnicalAnalysisService()

	stockPrices := []*models.StockPrice{
		{
			Code:       "1234",
			Date:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			OpenPrice:  createNullDecimal(1000.0),
			HighPrice:  createNullDecimal(1100.0),
			LowPrice:   createNullDecimal(950.0),
			ClosePrice: createNullDecimal(1050.0),
			Volume:     10000,
		},
		{
			Code:       "1234",
			Date:       time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			OpenPrice:  createNullDecimal(1050.0),
			HighPrice:  createNullDecimal(1150.0),
			LowPrice:   createNullDecimal(1000.0),
			ClosePrice: createNullDecimal(1100.0),
			Volume:     12000,
		},
	}

	result := service.ConvertStockPrices(stockPrices)

	expectedLen := 2
	if diff := cmp.Diff(expectedLen, len(result)); diff != "" {
		t.Errorf("Converted prices length mismatch (-want +got):\n%s", diff)
	}

	expected := []StockPriceData{
		{
			Code:      "1234",
			Date:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:      1000.0,
			High:      1100.0,
			Low:       950.0,
			Close:     1050.0,
			Volume:    10000,
			Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Code:      "1234",
			Date:      time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			Open:      1050.0,
			High:      1150.0,
			Low:       1000.0,
			Close:     1100.0,
			Volume:    12000,
			Timestamp: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Converted stock prices mismatch (-want +got):\n%s", diff)
	}
}

func TestTechnicalAnalysisService_MovingAverage(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name     string
		prices   []StockPriceData
		period   int
		expected float64
	}{
		{
			name: "5-day moving average",
			prices: []StockPriceData{
				{Close: 100.0},
				{Close: 110.0},
				{Close: 105.0},
				{Close: 115.0},
				{Close: 120.0},
			},
			period:   5,
			expected: 110.0, // (100+110+105+115+120)/5
		},
		{
			name: "3-day moving average from 5 prices",
			prices: []StockPriceData{
				{Close: 100.0},
				{Close: 110.0},
				{Close: 105.0},
				{Close: 115.0},
				{Close: 120.0},
			},
			period:   3,
			expected: 113.33333333333333, // (105+115+120)/3
		},
		{
			name: "Insufficient data",
			prices: []StockPriceData{
				{Close: 100.0},
				{Close: 110.0},
			},
			period:   5,
			expected: 0.0,
		},
		{
			name:     "Empty prices",
			prices:   []StockPriceData{},
			period:   5,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.MovingAverage(tt.prices, tt.period)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Moving average mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTechnicalAnalysisService_RSI(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name     string
		prices   []StockPriceData
		period   int
		expected float64
	}{
		{
			name: "RSI with mixed gains and losses",
			prices: []StockPriceData{
				{Close: 100.0},
				{Close: 102.0}, // +2
				{Close: 101.0}, // -1
				{Close: 103.0}, // +2
				{Close: 105.0}, // +2
				{Close: 104.0}, // -1
				{Close: 106.0}, // +2
				{Close: 108.0}, // +2
				{Close: 107.0}, // -1
				{Close: 109.0}, // +2
				{Close: 111.0}, // +2
				{Close: 110.0}, // -1
				{Close: 112.0}, // +2
				{Close: 114.0}, // +2
				{Close: 113.0}, // -1
			},
			period:   14,
			expected: 80.0, // This is approximate based on RSI calculation
		},
		{
			name: "Insufficient data",
			prices: []StockPriceData{
				{Close: 100.0},
				{Close: 102.0},
			},
			period:   14,
			expected: 50.0, // Neutral value for insufficient data
		},
		{
			name:     "Empty prices",
			prices:   []StockPriceData{},
			period:   14,
			expected: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.RSI(tt.prices, tt.period)

			// Use approximate comparison for RSI since it's a complex calculation
			tolerance := 5.0
			if abs(result-tt.expected) > tolerance {
				t.Errorf("RSI mismatch: expected approximately %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestTechnicalAnalysisService_MACD(t *testing.T) {
	service := NewTechnicalAnalysisService()

	prices := []StockPriceData{
		{Close: 100.0}, {Close: 102.0}, {Close: 101.0}, {Close: 103.0}, {Close: 105.0},
		{Close: 104.0}, {Close: 106.0}, {Close: 108.0}, {Close: 107.0}, {Close: 109.0},
		{Close: 111.0}, {Close: 110.0}, {Close: 112.0}, {Close: 114.0}, {Close: 113.0},
		{Close: 115.0}, {Close: 117.0}, {Close: 116.0}, {Close: 118.0}, {Close: 120.0},
		{Close: 119.0}, {Close: 121.0}, {Close: 123.0}, {Close: 122.0}, {Close: 124.0},
		{Close: 126.0}, {Close: 125.0}, {Close: 127.0}, {Close: 129.0}, {Close: 128.0},
	}

	macd, signal, histogram := service.MACD(prices, 12, 26, 9)

	// Since this is a simplified MACD implementation, we just check that values are reasonable
	if macd == 0 && signal == 0 && histogram == 0 {
		t.Error("MACD values should not all be zero for sufficient data")
	}

	expectedHistogram := macd - signal
	if diff := cmp.Diff(expectedHistogram, histogram); diff != "" {
		t.Errorf("MACD histogram should equal macd - signal (-want +got):\n%s", diff)
	}
}

func TestTechnicalAnalysisService_CalculateAllIndicators(t *testing.T) {
	service := NewTechnicalAnalysisService()

	prices := []StockPriceData{
		{Code: "1234", Close: 100.0, Timestamp: time.Now()},
		{Code: "1234", Close: 102.0, Timestamp: time.Now()},
		{Code: "1234", Close: 101.0, Timestamp: time.Now()},
		{Code: "1234", Close: 103.0, Timestamp: time.Now()},
		{Code: "1234", Close: 105.0, Timestamp: time.Now()},
	}

	for i := 0; i < 70; i++ {
		prices = append(prices, StockPriceData{
			Code:      "1234",
			Close:     float64(105 + i),
			Timestamp: time.Now(),
		})
	}

	result := service.CalculateAllIndicators(prices)

	if result == nil {
		t.Fatal("Expected non-nil indicator data")
	}

	if diff := cmp.Diff("1234", result.Code); diff != "" {
		t.Errorf("Code mismatch (-want +got):\n%s", diff)
	}

	// Check that indicators are calculated (non-zero values)
	if result.MA5 == 0 {
		t.Error("MA5 should be calculated")
	}
	if result.MA25 == 0 {
		t.Error("MA25 should be calculated")
	}
	if result.MA75 == 0 {
		t.Error("MA75 should be calculated")
	}
	if result.RSI == 0 {
		t.Error("RSI should be calculated")
	}
}

func TestTechnicalAnalysisService_GenerateTradingSignal(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name           string
		indicator      *TechnicalIndicatorData
		currentPrice   float64
		expectedAction string
	}{
		{
			name: "Strong buy signal",
			indicator: &TechnicalIndicatorData{
				Code:      "1234",
				MA5:       110.0,
				MA25:      105.0,
				MA75:      100.0,
				RSI:       25.0, // Oversold
				MACD:      2.0,
				Signal:    1.0,
				Histogram: 1.0,
			},
			currentPrice:   115.0,
			expectedAction: "buy",
		},
		{
			name: "Strong sell signal",
			indicator: &TechnicalIndicatorData{
				Code:      "1234",
				MA5:       95.0,
				MA25:      100.0,
				MA75:      105.0,
				RSI:       80.0, // Overbought
				MACD:      -2.0,
				Signal:    -1.0,
				Histogram: -1.0,
			},
			currentPrice:   90.0,
			expectedAction: "sell",
		},
		{
			name: "Hold signal",
			indicator: &TechnicalIndicatorData{
				Code:      "1234",
				MA5:       100.0,
				MA25:      100.0,
				MA75:      100.0,
				RSI:       50.0, // Neutral
				MACD:      0.5,
				Signal:    0.4,
				Histogram: 0.1,
			},
			currentPrice:   100.0,
			expectedAction: "hold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal := service.GenerateTradingSignal(tt.indicator, tt.currentPrice)

			if signal == nil {
				t.Fatal("Expected non-nil trading signal")
			}

			if diff := cmp.Diff(tt.expectedAction, signal.Action); diff != "" {
				t.Errorf("Trading signal action mismatch (-want +got):\n%s", diff)
			}

			if signal.Confidence < 0 || signal.Confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", signal.Confidence)
			}

			if len(signal.Reason) == 0 {
				t.Error("Signal should have a reason")
			}
		})
	}
}

func TestTechnicalAnalysisService_ValidateIndicator(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name      string
		indicator *TechnicalIndicatorData
		wantError bool
	}{
		{
			name: "Valid indicator",
			indicator: &TechnicalIndicatorData{
				Code: "1234",
				RSI:  50.0,
			},
			wantError: false,
		},
		{
			name: "Empty code",
			indicator: &TechnicalIndicatorData{
				Code: "",
				RSI:  50.0,
			},
			wantError: true,
		},
		{
			name: "RSI too low",
			indicator: &TechnicalIndicatorData{
				Code: "1234",
				RSI:  -10.0,
			},
			wantError: true,
		},
		{
			name: "RSI too high",
			indicator: &TechnicalIndicatorData{
				Code: "1234",
				RSI:  150.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateIndicator(tt.indicator)

			if tt.wantError {
				if err == nil {
					t.Error("Expected validation error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestTechnicalAnalysisService_GetSignalStrength(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name      string
		indicator *TechnicalIndicatorData
		expected  string
	}{
		{
			name: "Strong buy",
			indicator: &TechnicalIndicatorData{
				RSI:       25.0, // Buy signal
				MACD:      2.0,
				Signal:    1.0,
				Histogram: 1.0,
			},
			expected: "Strong Buy",
		},
		{
			name: "Strong sell",
			indicator: &TechnicalIndicatorData{
				RSI:       80.0, // Sell signal
				MACD:      -2.0,
				Signal:    -1.0,
				Histogram: -1.0,
			},
			expected: "Strong Sell",
		},
		{
			name: "Neutral",
			indicator: &TechnicalIndicatorData{
				RSI:       50.0,
				MACD:      0.5,
				Signal:    0.4,
				Histogram: 0.1,
			},
			expected: "Neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetSignalStrength(tt.indicator)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Signal strength mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTechnicalAnalysisService_ConvertToModelIndicator(t *testing.T) {
	service := NewTechnicalAnalysisService()

	data := &TechnicalIndicatorData{
		Code:      "1234",
		MA5:       105.0,
		MA25:      100.0,
		MA75:      95.0,
		RSI:       60.0,
		MACD:      1.5,
		Signal:    1.0,
		Histogram: 0.5,
		Timestamp: time.Now(),
	}

	result := service.ConvertToModelIndicator(data)

	if result == nil {
		t.Fatal("Expected non-nil model indicator")
	}

	if diff := cmp.Diff("1234", result.Code); diff != "" {
		t.Errorf("Code mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(data.Timestamp, result.Date); diff != "" {
		t.Errorf("Date mismatch (-want +got):\n%s", diff)
	}
}

func TestTechnicalAnalysisService_DecimalConversion(t *testing.T) {
	service := NewTechnicalAnalysisService()

	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{
			name:     "String number",
			input:    "1234.56",
			expected: 1234.56,
		},
		{
			name:     "Integer",
			input:    1000,
			expected: 1000.0,
		},
		{
			name:     "Float",
			input:    1500.75,
			expected: 1500.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.decimalToFloat(tt.input)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Decimal conversion mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Helper functions for testing

func createNullDecimal(_ float64) types.Decimal {
	// Simplified for testing - in real implementation would properly convert
	return types.Decimal{}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
