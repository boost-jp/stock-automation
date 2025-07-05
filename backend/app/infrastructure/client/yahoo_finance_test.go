package client

import (
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
)

func TestYahooFinanceClient_Interface(t *testing.T) {
	client := NewYahooFinanceClient()

	// Verify interface compliance
	var _ StockDataClient = client

	if client == nil {
		t.Error("Yahoo Finance client should not be nil")
	}

	if client.baseURL == "" {
		t.Error("Base URL should be set")
	}

	if client.client == nil {
		t.Error("HTTP client should be initialized")
	}
}

func TestYahooFinanceClient_GetCurrentPrice_InvalidCode(t *testing.T) {
	client := NewYahooFinanceClient()

	// Test with invalid stock code
	_, err := client.GetCurrentPrice("INVALID_CODE_12345")
	if err == nil {
		t.Error("Expected error for invalid stock code")
	}
}

func TestYahooFinanceClient_GetHistoricalData_InvalidParams(t *testing.T) {
	client := NewYahooFinanceClient()

	tests := []struct {
		name      string
		stockCode string
		days      int
		wantError bool
	}{
		{
			name:      "Invalid stock code",
			stockCode: "INVALID_CODE_12345",
			days:      30,
			wantError: true,
		},
		{
			name:      "Zero days",
			stockCode: "1234",
			days:      0,
			wantError: false, // Should handle gracefully
		},
		{
			name:      "Negative days",
			stockCode: "1234",
			days:      -30,
			wantError: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetHistoricalData(tt.stockCode, tt.days)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}

			// Note: We can't test successful cases without real API calls
			// In integration tests, you would test with real stock codes
		})
	}
}

func TestYahooFinanceClient_GetIntradayData_InvalidParams(t *testing.T) {
	client := NewYahooFinanceClient()

	tests := []struct {
		name      string
		stockCode string
		interval  string
		wantError bool
	}{
		{
			name:      "Invalid stock code",
			stockCode: "INVALID_CODE_12345",
			interval:  "1m",
			wantError: true,
		},
		{
			name:      "Valid parameters",
			stockCode: "1234",
			interval:  "5m",
			wantError: true, // Will fail without real API access
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetIntradayData(tt.stockCode, tt.interval)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// MockStockDataClient implements StockDataClient for testing.
type MockStockDataClient struct {
	mockCurrentPrice   *models.StockPrice
	mockHistoricalData []*models.StockPrice
	mockIntradayData   []*models.StockPrice
	shouldReturnError  bool
}

func NewMockStockDataClient() *MockStockDataClient {
	return &MockStockDataClient{
		mockCurrentPrice: &models.StockPrice{
			Code:       "1234",
			Date:       time.Now(),
			OpenPrice:  floatToDecimal(1000.0),
			HighPrice:  floatToDecimal(1100.0),
			LowPrice:   floatToDecimal(950.0),
			ClosePrice: floatToDecimal(1050.0),
			Volume:     1000000,
		},
		mockHistoricalData: []*models.StockPrice{
			{
				Code:       "1234",
				Date:       time.Now().AddDate(0, 0, -1),
				OpenPrice:  floatToDecimal(1000.0),
				HighPrice:  floatToDecimal(1100.0),
				LowPrice:   floatToDecimal(950.0),
				ClosePrice: floatToDecimal(1050.0),
				Volume:     1000000,
			},
		},
		mockIntradayData: []*models.StockPrice{
			{
				Code:       "1234",
				Date:       time.Now().Add(-1 * time.Hour),
				OpenPrice:  floatToDecimal(1000.0),
				HighPrice:  floatToDecimal(1100.0),
				LowPrice:   floatToDecimal(950.0),
				ClosePrice: floatToDecimal(1050.0),
				Volume:     500000,
			},
		},
	}
}

func (m *MockStockDataClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	if m.shouldReturnError {
		return nil, &mockError{message: "mock error"}
	}
	return m.mockCurrentPrice, nil
}

func (m *MockStockDataClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	if m.shouldReturnError {
		return nil, &mockError{message: "mock error"}
	}
	return m.mockHistoricalData, nil
}

func (m *MockStockDataClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	if m.shouldReturnError {
		return nil, &mockError{message: "mock error"}
	}
	return m.mockIntradayData, nil
}

// mockError implements error interface for testing.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestMockStockDataClient(t *testing.T) {
	client := NewMockStockDataClient()

	// Test interface compliance
	var _ StockDataClient = client

	t.Run("GetCurrentPrice", func(t *testing.T) {
		price, err := client.GetCurrentPrice("1234")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if price == nil {
			t.Error("Price should not be nil")
		}
		if price.Code != "1234" {
			t.Errorf("Expected code 1234, got %s", price.Code)
		}
	})

	t.Run("GetHistoricalData", func(t *testing.T) {
		prices, err := client.GetHistoricalData("1234", 30)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(prices) == 0 {
			t.Error("Should return at least one price")
		}
	})

	t.Run("GetIntradayData", func(t *testing.T) {
		prices, err := client.GetIntradayData("1234", "1m")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(prices) == 0 {
			t.Error("Should return at least one price")
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		client.shouldReturnError = true

		_, err := client.GetCurrentPrice("1234")
		if err == nil {
			t.Error("Expected error but got none")
		}

		_, err = client.GetHistoricalData("1234", 30)
		if err == nil {
			t.Error("Expected error but got none")
		}

		_, err = client.GetIntradayData("1234", "1m")
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}
