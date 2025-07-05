package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
)

func TestDefaultYahooFinanceConfig(t *testing.T) {
	config := DefaultYahooFinanceConfig()

	if config.BaseURL != "https://query1.finance.yahoo.com" {
		t.Errorf("Expected base URL to be https://query1.finance.yahoo.com, got %s", config.BaseURL)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", config.Timeout)
	}
	if config.RetryCount != 3 {
		t.Errorf("Expected retry count to be 3, got %d", config.RetryCount)
	}
	if config.RateLimitRPS != 10 {
		t.Errorf("Expected rate limit to be 10 RPS, got %d", config.RateLimitRPS)
	}
}

func TestNewYahooFinanceClient(t *testing.T) {
	client := NewYahooFinanceClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.client == nil {
		t.Fatal("Expected non-nil HTTP client")
	}
	if client.rateLimiter == nil {
		t.Fatal("Expected non-nil rate limiter")
	}
}

func TestNewYahooFinanceClientWithConfig(t *testing.T) {
	config := YahooFinanceConfig{
		BaseURL:       "https://test.example.com",
		Timeout:       10 * time.Second,
		RetryCount:    5,
		RetryWaitTime: 2 * time.Second,
		RetryMaxWait:  20 * time.Second,
		UserAgent:     "TestAgent/1.0",
		RateLimitRPS:  5,
	}

	client := NewYahooFinanceClientWithConfig(config)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.baseURL != config.BaseURL {
		t.Errorf("Expected base URL %s, got %s", config.BaseURL, client.baseURL)
	}
	if client.rateLimiter == nil {
		t.Fatal("Expected non-nil rate limiter")
	}
}

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

func TestYahooFinanceClient_GetCurrentPrice_RateLimit(t *testing.T) {
	// Create a test server that always returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "TEST",
						"regularMarketPrice": 100.5,
						"previousClose": 99.0,
						"regularMarketOpen": 99.5,
						"regularMarketDayLow": 98.0,
						"regularMarketDayHigh": 101.0,
						"regularMarketVolume": 1000000
					}
				}]
			}
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := YahooFinanceConfig{
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		RetryCount:    1,
		RetryWaitTime: 100 * time.Millisecond,
		RetryMaxWait:  500 * time.Millisecond,
		RateLimitRPS:  2, // Low rate limit for testing
	}

	client := NewYahooFinanceClientWithConfig(config)

	// Make rapid requests to test rate limiting
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.GetCurrentPrice("TEST")
		if err != nil {
			t.Errorf("GetCurrentPrice() error = %v", err)
		}
	}
	elapsed := time.Since(start)

	// With 2 RPS and 3 requests, at least one should be delayed
	if elapsed < 400*time.Millisecond {
		t.Errorf("Expected rate limiting to delay requests, but completed in %v", elapsed)
	}
}

func TestYahooFinanceClient_GetCurrentPrice_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
		checkError func(error) bool
	}{
		{
			name:       "404 Not Found",
			statusCode: 404,
			wantErr:    true,
			checkError: func(err error) bool {
				return err != nil && IsRetryableError(err) == false
			},
		},
		{
			name:       "429 Rate Limit",
			statusCode: 429,
			wantErr:    true,
			checkError: func(err error) bool {
				return err != nil && IsRetryableError(err) == true
			},
		},
		{
			name:       "500 Server Error",
			statusCode: 500,
			wantErr:    true,
			checkError: func(err error) bool {
				return err != nil && IsRetryableError(err) == true
			},
		},
		{
			name:       "503 Service Unavailable",
			statusCode: 503,
			wantErr:    true,
			checkError: func(err error) bool {
				return err != nil && IsRetryableError(err) == true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			config := YahooFinanceConfig{
				BaseURL:       server.URL,
				Timeout:       1 * time.Second,
				RetryCount:    0, // Disable retries for predictable testing
				RateLimitRPS:  10,
			}

			client := NewYahooFinanceClientWithConfig(config)
			_, err := client.GetCurrentPrice("TEST")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentPrice() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkError != nil && !tt.checkError(err) {
				t.Errorf("Error check failed for error: %v", err)
			}
		})
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
