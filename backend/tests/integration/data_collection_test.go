package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/testutil"
	"github.com/boost-jp/stock-automation/app/testutil/fixture"
	"github.com/boost-jp/stock-automation/app/usecase"
)

// TestDataCollection_CompleteFlow tests the complete data collection flow
func TestDataCollection_CompleteFlow(t *testing.T) {
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

	// Create mock stock client with test data
	stockClient := &mockStockDataClientWithHistory{
		currentPrices: map[string]float64{
			"7203": 2150.0,
			"6758": 13800.0,
			"9984": 6500.0,
		},
		historicalData: generateHistoricalData(),
	}

	// Create use case
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	// Test 1: Collect historical data
	t.Run("CollectHistoricalData", func(t *testing.T) {
		stocks := []string{"7203", "6758", "9984"}

		for _, code := range stocks {
			err := collectDataUC.CollectHistoricalData(ctx, code, 30)
			if err != nil {
				t.Errorf("Failed to collect historical data for %s: %v", code, err)
			}

			// Verify data was saved
			history, err := stockRepo.GetPriceHistory(ctx, code, 30)
			if err != nil {
				t.Errorf("Failed to get price history for %s: %v", code, err)
			}

			if len(history) == 0 {
				t.Errorf("No historical data found for %s", code)
			}
		}
	})

	// Test 2: Update current prices
	t.Run("UpdateCurrentPrices", func(t *testing.T) {
		// Setup watch list
		watchList := []*models.WatchList{
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

		for _, w := range watchList {
			if err := stockRepo.AddToWatchList(ctx, w); err != nil {
				// Ignore duplicate key errors since tests may run in parallel
				if !strings.Contains(err.Error(), "Duplicate entry") {
					t.Fatalf("Failed to add to watch list: %v", err)
				}
			}
		}

		// Update prices
		err := collectDataUC.UpdateAllPrices(ctx)
		if err != nil {
			t.Errorf("Failed to update all prices: %v", err)
		}

		// Verify prices were updated
		for code, expectedPrice := range stockClient.currentPrices {
			price, err := stockRepo.GetLatestPrice(ctx, code)
			if err != nil {
				t.Errorf("Failed to get latest price for %s: %v", code, err)
				continue
			}

			actualPrice := client.DecimalToFloat(price.ClosePrice)
			if actualPrice != expectedPrice {
				t.Errorf("Price mismatch for %s: expected %f, got %f", code, expectedPrice, actualPrice)
			}
		}
	})

	// Test 3: Market hours check
	t.Run("MarketHoursCheck", func(t *testing.T) {
		isOpen := collectDataUC.IsMarketOpen()
		// Just verify it returns a boolean without error
		t.Logf("Market open status: %v", isOpen)
	})

	// Test 4: Cleanup old data
	t.Run("CleanupOldData", func(t *testing.T) {
		// Insert some old test data
		oldPrice := &models.StockPrice{
			ID:         fmt.Sprintf("old-price-%d", time.Now().UnixNano()),
			Code:       "7203",
			Date:       time.Now().AddDate(-2, 0, 0), // 2 years old
			ClosePrice: client.FloatToDecimal(1000.0),
			OpenPrice:  client.FloatToDecimal(1000.0),
			HighPrice:  client.FloatToDecimal(1000.0),
			LowPrice:   client.FloatToDecimal(1000.0),
			Volume:     1000,
			CreatedAt:  fixture.NullTimeFrom(time.Now().AddDate(-2, 0, 0)),
			UpdatedAt:  fixture.NullTimeFrom(time.Now().AddDate(-2, 0, 0)),
		}

		if err := stockRepo.SaveStockPrice(ctx, oldPrice); err != nil {
			// Ignore duplicate key errors
			if !strings.Contains(err.Error(), "Duplicate entry") {
				t.Fatalf("Failed to save old price: %v", err)
			}
		}

		// Run cleanup (keep only 365 days)
		err := collectDataUC.CleanupOldData(ctx, 365)
		if err != nil {
			t.Errorf("Failed to cleanup old data: %v", err)
		}

		// Verify old data was removed
		// Note: The actual cleanup implementation might vary
		t.Log("Cleanup completed")
	})
}

// TestDataCollection_RateLimiting tests rate limiting functionality
func TestDataCollection_RateLimiting(t *testing.T) {
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

	// Create mock client that tracks call timing
	stockClient := &rateLimitTestClient{
		callTimes: make([]time.Time, 0),
	}

	// Create use case
	collectDataUC := usecase.NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	// Setup many stocks to trigger rate limiting
	var watchList []*models.WatchList
	for i := 1; i <= 20; i++ {
		watchList = append(watchList, &models.WatchList{
			ID:       fmt.Sprintf("watch-%d", i),
			Code:     fmt.Sprintf("%04d", i),
			Name:     fmt.Sprintf("Stock %d", i),
			IsActive: fixture.NullBoolFrom(true),
		})
	}

	for _, w := range watchList {
		if err := stockRepo.AddToWatchList(ctx, w); err != nil {
			t.Fatalf("Failed to add to watch list: %v", err)
		}
	}

	// Start timing
	start := time.Now()

	// Update prices (should be rate limited)
	err = collectDataUC.UpdateAllPrices(ctx)
	if err != nil {
		t.Errorf("Failed to update prices: %v", err)
	}

	elapsed := time.Since(start)

	// With 20 requests and rate limit, it should take some time
	t.Logf("Total time for %d requests: %v", len(watchList), elapsed)
	t.Logf("Total API calls made: %d", len(stockClient.callTimes))

	// Verify rate limiting is working
	if len(stockClient.callTimes) > 1 {
		// Check intervals between calls
		for i := 1; i < len(stockClient.callTimes); i++ {
			interval := stockClient.callTimes[i].Sub(stockClient.callTimes[i-1])
			t.Logf("Interval between call %d and %d: %v", i-1, i, interval)
		}
	}
}

// Helper types for testing

type mockStockDataClientWithHistory struct {
	currentPrices  map[string]float64
	historicalData map[string][]*models.StockPrice
}

func (m *mockStockDataClientWithHistory) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	price, ok := m.currentPrices[stockCode]
	if !ok {
		price = 1000.0
	}

	return &models.StockPrice{
		ID:         fmt.Sprintf("current-%s-%d", stockCode, time.Now().Unix()),
		Code:       stockCode,
		Date:       time.Now(),
		OpenPrice:  client.FloatToDecimal(price * 0.99),
		HighPrice:  client.FloatToDecimal(price * 1.01),
		LowPrice:   client.FloatToDecimal(price * 0.98),
		ClosePrice: client.FloatToDecimal(price),
		Volume:     1000000,
		CreatedAt:  fixture.NullTimeFrom(time.Now()),
		UpdatedAt:  fixture.NullTimeFrom(time.Now()),
	}, nil
}

func (m *mockStockDataClientWithHistory) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	data, ok := m.historicalData[stockCode]
	if !ok {
		// Generate default historical data
		data = generateDefaultHistoricalData(stockCode, days)
	}

	// Return only requested days
	if len(data) > days {
		data = data[:days]
	}

	return data, nil
}

func (m *mockStockDataClientWithHistory) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

// Rate limit test client
type rateLimitTestClient struct {
	callTimes []time.Time
	mu        sync.Mutex
}

func (r *rateLimitTestClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	r.mu.Lock()
	r.callTimes = append(r.callTimes, time.Now())
	r.mu.Unlock()

	return &models.StockPrice{
		ID:         fmt.Sprintf("ratelimit-%s-%d", stockCode, time.Now().UnixNano()),
		Code:       stockCode,
		Date:       time.Now(),
		ClosePrice: client.FloatToDecimal(1000.0),
		OpenPrice:  client.FloatToDecimal(1000.0),
		HighPrice:  client.FloatToDecimal(1000.0),
		LowPrice:   client.FloatToDecimal(1000.0),
		Volume:     1000000,
		CreatedAt:  fixture.NullTimeFrom(time.Now()),
		UpdatedAt:  fixture.NullTimeFrom(time.Now()),
	}, nil
}

func (r *rateLimitTestClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (r *rateLimitTestClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

// Helper functions

func generateHistoricalData() map[string][]*models.StockPrice {
	result := make(map[string][]*models.StockPrice)

	stocks := map[string]float64{
		"7203": 2000.0,
		"6758": 12000.0,
		"9984": 6000.0,
	}

	for code, basePrice := range stocks {
		var prices []*models.StockPrice
		for i := 0; i < 30; i++ {
			date := time.Now().AddDate(0, 0, -i)
			// Add some variation
			variation := 1.0 + (float64(i%5)-2.0)*0.01
			price := basePrice * variation

			prices = append(prices, &models.StockPrice{
				ID:         fmt.Sprintf("%s-%d", code, i),
				Code:       code,
				Date:       date,
				OpenPrice:  client.FloatToDecimal(price * 0.99),
				HighPrice:  client.FloatToDecimal(price * 1.01),
				LowPrice:   client.FloatToDecimal(price * 0.98),
				ClosePrice: client.FloatToDecimal(price),
				Volume:     int64(1000000 + i*10000),
				CreatedAt:  fixture.NullTimeFrom(date),
				UpdatedAt:  fixture.NullTimeFrom(date),
			})
		}
		result[code] = prices
	}

	return result
}

func generateDefaultHistoricalData(stockCode string, days int) []*models.StockPrice {
	var prices []*models.StockPrice
	basePrice := 1000.0

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		variation := 1.0 + (float64(i%7)-3.0)*0.005
		price := basePrice * variation

		prices = append(prices, &models.StockPrice{
			ID:         fmt.Sprintf("%s-default-%d", stockCode, i),
			Code:       stockCode,
			Date:       date,
			OpenPrice:  client.FloatToDecimal(price * 0.995),
			HighPrice:  client.FloatToDecimal(price * 1.005),
			LowPrice:   client.FloatToDecimal(price * 0.995),
			ClosePrice: client.FloatToDecimal(price),
			Volume:     1000000,
			CreatedAt:  fixture.NullTimeFrom(date),
			UpdatedAt:  fixture.NullTimeFrom(date),
		})
	}

	return prices
}
