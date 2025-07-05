package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/sirupsen/logrus"
)

// CollectDataUseCase handles data collection business logic.
type CollectDataUseCase struct {
	stockRepo    repository.StockRepository
	portfolioRepo repository.PortfolioRepository
	stockClient  client.StockDataClient
	mu           sync.RWMutex
	watchList    []*models.WatchList
	portfolio    []*models.Portfolio
}

// NewCollectDataUseCase creates a new data collection use case.
func NewCollectDataUseCase(
	stockRepo repository.StockRepository,
	portfolioRepo repository.PortfolioRepository,
	stockClient client.StockDataClient,
) *CollectDataUseCase {
	return &CollectDataUseCase{
		stockRepo:     stockRepo,
		portfolioRepo: portfolioRepo,
		stockClient:   stockClient,
	}
}

// UpdateWatchList updates the watch list from database.
func (uc *CollectDataUseCase) UpdateWatchList(ctx context.Context) error {
	watchList, err := uc.stockRepo.GetActiveWatchList(ctx)
	if err != nil {
		return err
	}

	uc.mu.Lock()
	uc.watchList = watchList
	uc.mu.Unlock()

	logrus.Infof("Watch list updated: %d stocks", len(watchList))
	return nil
}

// UpdatePortfolio updates the portfolio from database.
func (uc *CollectDataUseCase) UpdatePortfolio(ctx context.Context) error {
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	uc.mu.Lock()
	uc.portfolio = portfolio
	uc.mu.Unlock()

	logrus.Infof("Portfolio updated: %d holdings", len(portfolio))
	return nil
}

// UpdateAllPrices updates prices for all watched stocks and portfolio.
func (uc *CollectDataUseCase) UpdateAllPrices(ctx context.Context) error {
	uc.mu.RLock()
	watchList := uc.watchList
	portfolio := uc.portfolio
	uc.mu.RUnlock()

	// Collect all unique stock codes
	stockCodes := make(map[string]bool)
	for _, item := range watchList {
		stockCodes[item.Code] = true
	}
	for _, item := range portfolio {
		stockCodes[item.Code] = true
	}

	// Update prices for all stocks
	var wg sync.WaitGroup
	for code := range stockCodes {
		wg.Add(1)
		go func(stockCode string) {
			defer wg.Done()
			if err := uc.UpdateStockPrice(ctx, stockCode); err != nil {
				logrus.Errorf("Failed to update price for %s: %v", stockCode, err)
			}
		}(code)
	}

	wg.Wait()
	return nil
}

// UpdateStockPrice updates the price for a single stock.
func (uc *CollectDataUseCase) UpdateStockPrice(ctx context.Context, stockCode string) error {
	price, err := uc.stockClient.GetCurrentPrice(stockCode)
	if err != nil {
		return err
	}

	if err := uc.stockRepo.SaveStockPrice(ctx, price); err != nil {
		return err
	}

	logrus.Debugf("Price updated for %s: %.2f", stockCode, price.ClosePrice)
	return nil
}

// CollectHistoricalData collects historical data for technical analysis.
func (uc *CollectDataUseCase) CollectHistoricalData(ctx context.Context, stockCode string, days int) error {
	prices, err := uc.stockClient.GetHistoricalData(stockCode, days)
	if err != nil {
		return err
	}

	if err := uc.stockRepo.SaveStockPrices(ctx, prices); err != nil {
		return err
	}

	logrus.Infof("Historical data collected for %s: %d records", stockCode, len(prices))
	return nil
}

// IsMarketOpen checks if the market is currently open.
func (uc *CollectDataUseCase) IsMarketOpen() bool {
	now := time.Now()
	jst, _ := time.LoadLocation("Asia/Tokyo")
	nowJST := now.In(jst)

	// Check if weekend
	if nowJST.Weekday() == time.Saturday || nowJST.Weekday() == time.Sunday {
		return false
	}

	// Market hours: 9:00 - 15:00 JST
	hour := nowJST.Hour()
	return hour >= 9 && hour < 15
}

// CleanupOldData removes old data from the database.
func (uc *CollectDataUseCase) CleanupOldData(ctx context.Context, days int) error {
	return uc.stockRepo.CleanupOldData(ctx, days)
}