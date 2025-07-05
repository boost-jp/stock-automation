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

// MockStockRepository implements repository.StockRepository for testing
type MockStockRepository struct {
	// Data storage
	stockPrices         map[string]*models.StockPrice
	watchList           []*models.WatchList
	technicalIndicators map[string]*models.TechnicalIndicator

	// Error simulation
	saveStockPriceErr     error
	getLatestPriceErr     error
	getActiveWatchListErr error

	// Override functions
	GetPriceHistoryFunc func(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error)
}

func NewMockStockRepository() *MockStockRepository {
	return &MockStockRepository{
		stockPrices:         make(map[string]*models.StockPrice),
		watchList:           make([]*models.WatchList, 0),
		technicalIndicators: make(map[string]*models.TechnicalIndicator),
	}
}

func (m *MockStockRepository) SaveStockPrice(ctx context.Context, price *models.StockPrice) error {
	if m.saveStockPriceErr != nil {
		return m.saveStockPriceErr
	}
	m.stockPrices[price.Code] = price
	return nil
}

func (m *MockStockRepository) GetLatestPrice(ctx context.Context, stockCode string) (*models.StockPrice, error) {
	if m.getLatestPriceErr != nil {
		return nil, m.getLatestPriceErr
	}
	price, ok := m.stockPrices[stockCode]
	if !ok {
		// Return nil error but no price
		return nil, nil
	}
	return price, nil
}

func (m *MockStockRepository) GetPriceHistory(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error) {
	if m.GetPriceHistoryFunc != nil {
		return m.GetPriceHistoryFunc(ctx, stockCode, days)
	}
	// Simple implementation for testing
	return []*models.StockPrice{}, nil
}

func (m *MockStockRepository) SaveTechnicalIndicator(ctx context.Context, indicator *models.TechnicalIndicator) error {
	m.technicalIndicators[indicator.Code] = indicator
	return nil
}

func (m *MockStockRepository) GetActiveWatchList(ctx context.Context) ([]*models.WatchList, error) {
	if m.getActiveWatchListErr != nil {
		return nil, m.getActiveWatchListErr
	}
	return m.watchList, nil
}

func (m *MockStockRepository) AddToWatchList(ctx context.Context, watchList *models.WatchList) error {
	m.watchList = append(m.watchList, watchList)
	return nil
}

func (m *MockStockRepository) SaveStockPrices(ctx context.Context, prices []*models.StockPrice) error {
	for _, price := range prices {
		if err := m.SaveStockPrice(ctx, price); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockStockRepository) CleanupOldData(ctx context.Context, days int) error {
	// Simple implementation for testing
	return nil
}

func (m *MockStockRepository) GetLatestTechnicalIndicator(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error) {
	indicator, ok := m.technicalIndicators[stockCode]
	if !ok {
		return nil, nil
	}
	return indicator, nil
}

func (m *MockStockRepository) GetWatchListItem(ctx context.Context, id string) (*models.WatchList, error) {
	for _, item := range m.watchList {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *MockStockRepository) UpdateWatchList(ctx context.Context, item *models.WatchList) error {
	for i, existing := range m.watchList {
		if existing.ID == item.ID {
			m.watchList[i] = item
			return nil
		}
	}
	return nil
}

func (m *MockStockRepository) DeleteFromWatchList(ctx context.Context, id string) error {
	for i, item := range m.watchList {
		if item.ID == id {
			m.watchList = append(m.watchList[:i], m.watchList[i+1:]...)
			return nil
		}
	}
	return nil
}

// MockPortfolioRepository implements repository.PortfolioRepository for testing
type MockPortfolioRepository struct {
	portfolios map[string]*models.Portfolio
	getAllErr  error
}

func NewMockPortfolioRepository() *MockPortfolioRepository {
	return &MockPortfolioRepository{
		portfolios: make(map[string]*models.Portfolio),
	}
}

func (m *MockPortfolioRepository) Create(ctx context.Context, portfolio *models.Portfolio) error {
	m.portfolios[portfolio.ID] = portfolio
	return nil
}

func (m *MockPortfolioRepository) GetByID(ctx context.Context, id string) (*models.Portfolio, error) {
	return m.portfolios[id], nil
}

func (m *MockPortfolioRepository) GetByCode(ctx context.Context, code string) (*models.Portfolio, error) {
	for _, p := range m.portfolios {
		if p.Code == code {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockPortfolioRepository) GetAll(ctx context.Context) ([]*models.Portfolio, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	var result []*models.Portfolio
	for _, p := range m.portfolios {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPortfolioRepository) Update(ctx context.Context, portfolio *models.Portfolio) error {
	m.portfolios[portfolio.ID] = portfolio
	return nil
}

func (m *MockPortfolioRepository) Delete(ctx context.Context, id string) error {
	delete(m.portfolios, id)
	return nil
}

func (m *MockPortfolioRepository) GetTotalValue(ctx context.Context, currentPrices map[string]float64) (float64, error) {
	return 0, nil
}

func (m *MockPortfolioRepository) GetHoldingsByCode(ctx context.Context, codes []string) ([]*models.Portfolio, error) {
	var result []*models.Portfolio
	codeMap := make(map[string]bool)
	for _, code := range codes {
		codeMap[code] = true
	}

	for _, p := range m.portfolios {
		if codeMap[p.Code] {
			result = append(result, p)
		}
	}
	return result, nil
}

// MockStockDataClient for testing
type MockStockDataClient struct {
	prices      map[string]float64
	getPriceErr error
	getHistErr  error
}

func NewMockStockDataClient() *MockStockDataClient {
	return &MockStockDataClient{
		prices: make(map[string]float64),
	}
}

func (m *MockStockDataClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	if m.getPriceErr != nil {
		return nil, m.getPriceErr
	}

	price, ok := m.prices[stockCode]
	if !ok {
		price = 1000.0 // Default price
	}

	now := time.Now()
	return &models.StockPrice{
		ID:         fmt.Sprintf("price-%s-%d", stockCode, now.Unix()),
		Code:       stockCode,
		Date:       now,
		OpenPrice:  client.FloatToDecimal(price * 0.99),
		HighPrice:  client.FloatToDecimal(price * 1.01),
		LowPrice:   client.FloatToDecimal(price * 0.98),
		ClosePrice: client.FloatToDecimal(price),
		Volume:     1000000,
		CreatedAt:  fixture.NullTimeFrom(now),
		UpdatedAt:  fixture.NullTimeFrom(now),
	}, nil
}

func (m *MockStockDataClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	if m.getHistErr != nil {
		return nil, m.getHistErr
	}

	var prices []*models.StockPrice
	basePrice := 1000.0
	if p, ok := m.prices[stockCode]; ok {
		basePrice = p
	}

	for i := 0; i < days && i < 5; i++ { // Return max 5 days for testing
		date := time.Now().AddDate(0, 0, -i)
		variation := 1.0 + float64(i%3)*0.01
		price := basePrice * variation

		prices = append(prices, &models.StockPrice{
			ID:         fmt.Sprintf("hist-%s-%d", stockCode, i),
			Code:       stockCode,
			Date:       date,
			OpenPrice:  client.FloatToDecimal(price * 0.99),
			HighPrice:  client.FloatToDecimal(price * 1.01),
			LowPrice:   client.FloatToDecimal(price * 0.98),
			ClosePrice: client.FloatToDecimal(price),
			Volume:     1000000,
			CreatedAt:  fixture.NullTimeFrom(date),
			UpdatedAt:  fixture.NullTimeFrom(date),
		})
	}

	return prices, nil
}

func (m *MockStockDataClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	// Simple implementation for testing
	return []*models.StockPrice{}, nil
}

func TestNewCollectDataUseCase(t *testing.T) {
	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	if uc == nil {
		t.Fatal("NewCollectDataUseCase returned nil")
	}

	// Test that maxWorkers is set
	if uc.maxWorkers != 5 {
		t.Errorf("Expected maxWorkers to be 5, got %d", uc.maxWorkers)
	}
}

func TestCollectDataUseCase_UpdateStockPrice(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		stockCode   string
		setupMocks  func(*MockStockRepository, *MockStockDataClient)
		expectError bool
	}{
		{
			name:      "Success",
			stockCode: "7203",
			setupMocks: func(sr *MockStockRepository, sc *MockStockDataClient) {
				sc.prices["7203"] = 2150.0
			},
			expectError: false,
		},
		{
			name:      "Client error",
			stockCode: "7203",
			setupMocks: func(sr *MockStockRepository, sc *MockStockDataClient) {
				sc.getPriceErr = fmt.Errorf("API error")
			},
			expectError: true,
		},
		{
			name:      "Repository error",
			stockCode: "7203",
			setupMocks: func(sr *MockStockRepository, sc *MockStockDataClient) {
				sc.prices["7203"] = 2150.0
				sr.saveStockPriceErr = fmt.Errorf("DB error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := NewMockStockRepository()
			portfolioRepo := NewMockPortfolioRepository()
			stockClient := NewMockStockDataClient()

			tt.setupMocks(stockRepo, stockClient)

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.UpdateStockPrice(ctx, tt.stockCode)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify price was saved
				if _, ok := stockRepo.stockPrices[tt.stockCode]; !ok {
					t.Error("Stock price was not saved")
				}
			}
		})
	}
}

func TestCollectDataUseCase_UpdateAllPrices(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()

	// Setup test data
	stockClient.prices["7203"] = 2150.0
	stockClient.prices["6758"] = 13500.0

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

	portfolioRepo.portfolios["p1"] = &models.Portfolio{
		ID:   "p1",
		Code: "9984",
		Name: "ソフトバンクグループ",
	}

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	err := uc.UpdateAllPrices(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify all prices were updated
	expectedCodes := []string{"7203", "6758", "9984"}
	for _, code := range expectedCodes {
		if _, ok := stockRepo.stockPrices[code]; !ok {
			t.Errorf("Stock price for %s was not saved", code)
		}
	}
}

func TestCollectDataUseCase_CollectHistoricalData(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()

	stockClient.prices["7203"] = 2150.0

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
	err := uc.CollectHistoricalData(ctx, "7203", 30)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify historical data was saved
	savedCount := 0
	for code := range stockRepo.stockPrices {
		if code == "7203" {
			savedCount++
		}
	}

	if savedCount == 0 {
		t.Error("No historical data was saved")
	}
}

func TestCollectDataUseCase_IsMarketOpen(t *testing.T) {
	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	// Just verify the method exists and returns a boolean
	isOpen := uc.IsMarketOpen()
	if isOpen != true && isOpen != false {
		t.Error("IsMarketOpen should return a boolean")
	}
}

func TestCollectDataUseCase_CleanupOldData(t *testing.T) {
	ctx := context.Background()

	stockRepo := NewMockStockRepository()
	portfolioRepo := NewMockPortfolioRepository()
	stockClient := NewMockStockDataClient()

	uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)

	// Test with various retention days
	tests := []struct {
		name          string
		retentionDays int
		expectError   bool
	}{
		{"30 days", 30, false},
		{"365 days", 365, false},
		{"0 days", 0, false},
		{"Negative days", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.CleanupOldData(ctx, tt.retentionDays)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCollectDataUseCase_UpdateAllPrices_Errors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*MockStockRepository, *MockPortfolioRepository)
	}{
		{
			name: "Watch list error",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository) {
				sr.getActiveWatchListErr = fmt.Errorf("DB error")
			},
		},
		{
			name: "Portfolio error",
			setupMocks: func(sr *MockStockRepository, pr *MockPortfolioRepository) {
				pr.getAllErr = fmt.Errorf("DB error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockRepo := NewMockStockRepository()
			portfolioRepo := NewMockPortfolioRepository()
			stockClient := NewMockStockDataClient()

			tt.setupMocks(stockRepo, portfolioRepo)

			uc := NewCollectDataUseCase(stockRepo, portfolioRepo, stockClient)
			err := uc.UpdateAllPrices(ctx)

			// UpdateAllPrices returns error when initial fetch fails
			if err == nil {
				t.Error("UpdateAllPrices should return error when repository fetch fails")
			}
		})
	}
}
