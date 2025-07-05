package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/aarondl/sqlboiler/v4/types"
)

// Mock implementations for testing

type mockStockRepository struct {
	watchList      []*models.WatchList
	portfolioErr   error
	watchListErr   error
	savePriceErr   error
	savePricesErr  error
	cleanupErr     error
	savedPrices    []*models.StockPrice
}

func (m *mockStockRepository) SaveStockPrice(ctx context.Context, price *models.StockPrice) error {
	if m.savePriceErr != nil {
		return m.savePriceErr
	}
	m.savedPrices = append(m.savedPrices, price)
	return nil
}

func (m *mockStockRepository) SaveStockPrices(ctx context.Context, prices []*models.StockPrice) error {
	if m.savePricesErr != nil {
		return m.savePricesErr
	}
	m.savedPrices = append(m.savedPrices, prices...)
	return nil
}

func (m *mockStockRepository) GetLatestPrice(ctx context.Context, stockCode string) (*models.StockPrice, error) {
	return nil, nil
}

func (m *mockStockRepository) GetPriceHistory(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error) {
	return nil, nil
}

func (m *mockStockRepository) SaveTechnicalIndicator(ctx context.Context, indicator *models.TechnicalIndicator) error {
	return nil
}

func (m *mockStockRepository) GetActiveWatchList(ctx context.Context) ([]*models.WatchList, error) {
	if m.watchListErr != nil {
		return nil, m.watchListErr
	}
	return m.watchList, nil
}

func (m *mockStockRepository) CleanupOldData(ctx context.Context, days int) error {
	return m.cleanupErr
}

func (m *mockStockRepository) AddToWatchList(ctx context.Context, watchList *models.WatchList) error {
	return nil
}

func (m *mockStockRepository) RemoveFromWatchList(ctx context.Context, code string) error {
	return nil
}

func (m *mockStockRepository) UpdateWatchList(ctx context.Context, watchList *models.WatchList) error {
	return nil
}

func (m *mockStockRepository) GetStockNames(ctx context.Context, codes []string) (map[string]string, error) {
	return nil, nil
}

func (m *mockStockRepository) GetLatestTechnicalIndicator(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error) {
	return nil, nil
}

func (m *mockStockRepository) DeleteFromWatchList(ctx context.Context, code string) error {
	return nil
}

type mockPortfolioRepository struct {
	portfolio    []*models.Portfolio
	portfolioErr error
}

func (m *mockPortfolioRepository) GetAll(ctx context.Context) ([]*models.Portfolio, error) {
	if m.portfolioErr != nil {
		return nil, m.portfolioErr
	}
	return m.portfolio, nil
}

func (m *mockPortfolioRepository) GetByCode(ctx context.Context, code string) (*models.Portfolio, error) {
	return nil, nil
}

func (m *mockPortfolioRepository) Save(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) Update(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) Delete(ctx context.Context, code string) error {
	return nil
}

func (m *mockPortfolioRepository) GetHoldingsByCode(ctx context.Context, codes []string) ([]*models.Portfolio, error) {
	return nil, nil
}

func (m *mockPortfolioRepository) Create(ctx context.Context, portfolio *models.Portfolio) error {
	return nil
}

func (m *mockPortfolioRepository) UpdateSharesAndAveragePrice(ctx context.Context, code string, shares int, averagePrice types.Decimal) error {
	return nil
}

func (m *mockPortfolioRepository) GetByID(ctx context.Context, id string) (*models.Portfolio, error) {
	return nil, nil
}

func (m *mockPortfolioRepository) GetTotalValue(ctx context.Context) (types.Decimal, error) {
	return types.Decimal{}, nil
}

type mockStockDataClient struct {
	currentPrice    *models.StockPrice
	historicalData  []*models.StockPrice
	currentPriceErr error
	historicalErr   error
	callCount       int
}

func (m *mockStockDataClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	m.callCount++
	if m.currentPriceErr != nil {
		return nil, m.currentPriceErr
	}
	if m.currentPrice != nil {
		// Return a copy with the requested stock code
		price := *m.currentPrice
		price.Code = stockCode
		return &price, nil
	}
	return &models.StockPrice{
		Code:       stockCode,
		Date:       time.Now(),
		ClosePrice: client.FloatToDecimal(100.0),
		OpenPrice:  client.FloatToDecimal(99.0),
		HighPrice:  client.FloatToDecimal(101.0),
		LowPrice:   client.FloatToDecimal(98.0),
		Volume:     1000000,
	}, nil
}

func (m *mockStockDataClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	if m.historicalErr != nil {
		return nil, m.historicalErr
	}
	return m.historicalData, nil
}

func (m *mockStockDataClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	return nil, nil
}

// Tests

func TestNewCollectDataUseCase(t *testing.T) {
	stockRepo := &mockStockRepository{}
	portfolioRepo := &mockPortfolioRepository{}
	stockClient := &mockStockDataClient{}

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	if uc == nil {
		t.Fatal("Expected non-nil use case")
	}
	if uc.stockRepo == nil {
		t.Error("Expected stock repository to be set")
	}
	if uc.portfolioRepo == nil {
		t.Error("Expected portfolio repository to be set")
	}
	if uc.stockClient == nil {
		t.Error("Expected stock client to be set")
	}
	if uc.maxWorkers != 5 {
		t.Errorf("Expected maxWorkers to be 5, got %d", uc.maxWorkers)
	}
}

func TestCollectDataUseCase_UpdateWatchList(t *testing.T) {
	uc := &CollectDataUseCase{}
	
	// UpdateWatchList is now a no-op
	err := uc.UpdateWatchList(context.Background())
	if err != nil {
		t.Errorf("UpdateWatchList() error = %v, want nil", err)
	}
}

func TestCollectDataUseCase_UpdatePortfolio(t *testing.T) {
	uc := &CollectDataUseCase{}
	
	// UpdatePortfolio is now a no-op
	err := uc.UpdatePortfolio(context.Background())
	if err != nil {
		t.Errorf("UpdatePortfolio() error = %v, want nil", err)
	}
}

func TestCollectDataUseCase_UpdateAllPrices(t *testing.T) {
	tests := []struct {
		name          string
		watchList     []*models.WatchList
		portfolio     []*models.Portfolio
		watchListErr  error
		portfolioErr  error
		clientErr     error
		wantErr       bool
	}{
		{
			name: "successful update with watch list and portfolio",
			watchList: []*models.WatchList{
				{Code: "7203", Name: "Toyota"},
				{Code: "6758", Name: "Sony"},
			},
			portfolio: []*models.Portfolio{
				{Code: "7203", Name: "Toyota", Shares: 100},
				{Code: "9984", Name: "SoftBank", Shares: 50},
			},
			wantErr: false,
		},
		{
			name:         "watch list error",
			watchListErr: errors.New("database error"),
			wantErr:      true,
		},
		{
			name:      "portfolio error",
			watchList: []*models.WatchList{{Code: "7203", Name: "Toyota"}},
			portfolioErr: errors.New("database error"),
			wantErr:      true,
		},
		{
			name:      "empty watch list and portfolio",
			watchList: []*models.WatchList{},
			portfolio: []*models.Portfolio{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := &mockStockRepository{
				watchList:    tt.watchList,
				watchListErr: tt.watchListErr,
			}
			portfolioRepo := &mockPortfolioRepository{
				portfolio:    tt.portfolio,
				portfolioErr: tt.portfolioErr,
			}
			stockClient := &mockStockDataClient{
				currentPriceErr: tt.clientErr,
			}

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.UpdateAllPrices(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAllPrices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectDataUseCase_UpdatePricesForStocks(t *testing.T) {
	tests := []struct {
		name         string
		watchList    []*models.WatchList
		portfolio    []*models.Portfolio
		maxWorkers   int
		clientErr    error
		expectedCalls int
	}{
		{
			name: "update prices for unique stocks",
			watchList: []*models.WatchList{
				{Code: "7203"},
				{Code: "6758"},
			},
			portfolio: []*models.Portfolio{
				{Code: "7203"}, // Duplicate
				{Code: "9984"},
			},
			maxWorkers:    5,
			expectedCalls: 3, // 7203, 6758, 9984 (unique)
		},
		{
			name:       "empty lists",
			watchList:  []*models.WatchList{},
			portfolio:  []*models.Portfolio{},
			maxWorkers: 5,
			expectedCalls: 0,
		},
		{
			name: "client errors are handled",
			watchList: []*models.WatchList{
				{Code: "7203"},
			},
			maxWorkers:    1,
			clientErr:     errors.New("API error"),
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := &mockStockRepository{}
			portfolioRepo := &mockPortfolioRepository{}
			stockClient := &mockStockDataClient{
				currentPriceErr: tt.clientErr,
			}

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			uc.maxWorkers = tt.maxWorkers

			err := uc.UpdatePricesForStocks(context.Background(), tt.watchList, tt.portfolio)
			
			// Should not return error even if individual updates fail
			if err != nil {
				t.Errorf("UpdatePricesForStocks() error = %v, want nil", err)
			}

			if stockClient.callCount != tt.expectedCalls {
				t.Errorf("Expected %d API calls, got %d", tt.expectedCalls, stockClient.callCount)
			}
		})
	}
}

func TestCollectDataUseCase_UpdateStockPrice(t *testing.T) {
	tests := []struct {
		name         string
		stockCode    string
		clientErr    error
		saveErr      error
		wantErr      bool
	}{
		{
			name:      "successful price update",
			stockCode: "7203",
			wantErr:   false,
		},
		{
			name:      "client error",
			stockCode: "7203",
			clientErr: errors.New("API error"),
			wantErr:   true,
		},
		{
			name:      "save error",
			stockCode: "7203",
			saveErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := &mockStockRepository{
				savePriceErr: tt.saveErr,
			}
			portfolioRepo := &mockPortfolioRepository{}
			stockClient := &mockStockDataClient{
				currentPriceErr: tt.clientErr,
			}

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.UpdateStockPrice(context.Background(), tt.stockCode)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStockPrice() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.clientErr == nil && len(stockRepo.savedPrices) != 1 {
				t.Error("Expected price to be saved")
			}
		})
	}
}

func TestCollectDataUseCase_CollectHistoricalData(t *testing.T) {
	tests := []struct {
		name          string
		stockCode     string
		days          int
		historicalData []*models.StockPrice
		clientErr     error
		saveErr       error
		wantErr       bool
	}{
		{
			name:      "successful historical data collection",
			stockCode: "7203",
			days:      30,
			historicalData: []*models.StockPrice{
				{Code: "7203", Date: time.Now().AddDate(0, 0, -1)},
				{Code: "7203", Date: time.Now().AddDate(0, 0, -2)},
			},
			wantErr: false,
		},
		{
			name:      "client error",
			stockCode: "7203",
			days:      30,
			clientErr: errors.New("API error"),
			wantErr:   true,
		},
		{
			name:      "save error",
			stockCode: "7203",
			days:      30,
			historicalData: []*models.StockPrice{{Code: "7203"}},
			saveErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := &mockStockRepository{
				savePricesErr: tt.saveErr,
			}
			portfolioRepo := &mockPortfolioRepository{}
			stockClient := &mockStockDataClient{
				historicalData: tt.historicalData,
				historicalErr:  tt.clientErr,
			}

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.CollectHistoricalData(context.Background(), tt.stockCode, tt.days)

			if (err != nil) != tt.wantErr {
				t.Errorf("CollectHistoricalData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.clientErr == nil && len(stockRepo.savedPrices) != len(tt.historicalData) {
				t.Errorf("Expected %d prices to be saved, got %d", len(tt.historicalData), len(stockRepo.savedPrices))
			}
		})
	}
}

func TestCollectDataUseCase_IsMarketOpen(t *testing.T) {
	uc := &CollectDataUseCase{}

	// Note: This test will have different results depending on when it's run
	// For a more reliable test, we would need to mock time.Now()
	
	isOpen := uc.IsMarketOpen()
	// Just verify it returns a boolean without error
	_ = isOpen
}

func TestCollectDataUseCase_CleanupOldData(t *testing.T) {
	tests := []struct {
		name       string
		days       int
		cleanupErr error
		wantErr    bool
	}{
		{
			name:    "successful cleanup",
			days:    365,
			wantErr: false,
		},
		{
			name:       "cleanup error",
			days:       365,
			cleanupErr: errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := &mockStockRepository{
				cleanupErr: tt.cleanupErr,
			}
			portfolioRepo := &mockPortfolioRepository{}
			stockClient := &mockStockDataClient{}

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.CleanupOldData(context.Background(), tt.days)

			if (err != nil) != tt.wantErr {
				t.Errorf("CleanupOldData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectDataUseCase_ConcurrentUpdates(t *testing.T) {
	// Test that concurrent updates are properly limited by maxWorkers
	stockRepo := &mockStockRepository{}
	portfolioRepo := &mockPortfolioRepository{}
	
	// Track concurrent executions
	var maxConcurrent int
	var currentConcurrent int
	var mu sync.Mutex
	
	stockClient := &mockStockDataClient{
		currentPrice: &models.StockPrice{
			ClosePrice: client.FloatToDecimal(100),
		},
	}
	
	// Override GetCurrentPrice to track concurrency
	originalGetCurrentPrice := stockClient.GetCurrentPrice
	stockClient.GetCurrentPrice = func(stockCode string) (*models.StockPrice, error) {
		mu.Lock()
		currentConcurrent++
		if currentConcurrent > maxConcurrent {
			maxConcurrent = currentConcurrent
		}
		mu.Unlock()
		
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		
		result, err := originalGetCurrentPrice(stockCode)
		
		mu.Lock()
		currentConcurrent--
		mu.Unlock()
		
		return result, err
	}

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	uc.maxWorkers = 3 // Set low limit to test

	// Create many stocks to update
	var watchList []*models.WatchList
	for i := 0; i < 20; i++ {
		watchList = append(watchList, &models.WatchList{
			Code: fmt.Sprintf("TEST%d", i),
		})
	}

	ctx := context.Background()
	err := uc.UpdatePricesForStocks(ctx, watchList, nil)
	
	if err != nil {
		t.Errorf("UpdatePricesForStocks() error = %v", err)
	}

	if maxConcurrent > uc.maxWorkers {
		t.Errorf("Max concurrent workers (%d) exceeded limit (%d)", maxConcurrent, uc.maxWorkers)
	}
}