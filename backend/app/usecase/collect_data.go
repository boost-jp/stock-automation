package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/alert"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/sirupsen/logrus"
)

// CollectDataUseCase handles data collection business logic.
type CollectDataUseCase struct {
	stockRepo     repository.StockRepository
	portfolioRepo repository.PortfolioRepository
	stockClient   client.StockDataClient
	alertService  alert.Service
	maxWorkers    int
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
		maxWorkers:    5, // Limit concurrent API calls
	}
}

// SetAlertService sets the alert service for error notifications
func (uc *CollectDataUseCase) SetAlertService(alertService alert.Service) {
	uc.alertService = alertService
}

// UpdateWatchList is kept for backward compatibility but now is a no-op.
// Watch list is always fetched from database when needed.
func (uc *CollectDataUseCase) UpdateWatchList(ctx context.Context) error {
	// This is now a no-op as we fetch from DB on demand
	logrus.Info("UpdateWatchList called - data will be fetched from DB on demand")
	return nil
}

// UpdatePortfolio is kept for backward compatibility but now is a no-op.
// Portfolio is always fetched from database when needed.
func (uc *CollectDataUseCase) UpdatePortfolio(ctx context.Context) error {
	// This is now a no-op as we fetch from DB on demand
	logrus.Info("UpdatePortfolio called - data will be fetched from DB on demand")
	return nil
}

// UpdateAllPrices updates prices for all watched stocks and portfolio.
func (uc *CollectDataUseCase) UpdateAllPrices(ctx context.Context) error {
	// Fetch watch list from database
	watchList, err := uc.stockRepo.GetActiveWatchList(ctx)
	if err != nil {
		return err
	}

	// Fetch portfolio from database
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	return uc.UpdatePricesForStocks(ctx, watchList, portfolio)
}

// UpdatePricesForStocks updates prices for specific watch list and portfolio items.
func (uc *CollectDataUseCase) UpdatePricesForStocks(ctx context.Context, watchList []*models.WatchList, portfolio []*models.Portfolio) error {
	// Collect all unique stock codes
	stockCodes := make(map[string]bool)
	for _, item := range watchList {
		stockCodes[item.Code] = true
	}
	for _, item := range portfolio {
		stockCodes[item.Code] = true
	}

	// Create a channel for stock codes and a semaphore for limiting concurrency
	codeChan := make(chan string, len(stockCodes))
	for code := range stockCodes {
		codeChan <- code
	}
	close(codeChan)

	// Create worker pool
	var wg sync.WaitGroup
	errorChan := make(chan error, len(stockCodes))

	for i := 0; i < uc.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for stockCode := range codeChan {
				if err := uc.UpdateStockPrice(ctx, stockCode); err != nil {
					logrus.Errorf("Failed to update price for %s: %v", stockCode, err)
					errorChan <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errorChan)

	// Check if there were any errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		logrus.Warnf("Encountered %d errors during price updates", len(errors))

		// Send alert if too many errors
		if uc.alertService != nil && len(errors) > len(stockCodes)/2 {
			uc.alertService.SendError(ctx,
				"High Error Rate in Price Updates",
				fmt.Sprintf("Failed to update prices for %d out of %d stocks", len(errors), len(stockCodes)),
				fmt.Errorf("multiple price update failures: %d errors", len(errors)))
		}
	}

	return nil
}

// UpdateStockPrice updates the price for a single stock.
func (uc *CollectDataUseCase) UpdateStockPrice(ctx context.Context, stockCode string) error {
	price, err := uc.stockClient.GetCurrentPrice(stockCode)
	if err != nil {
		// Check if this is a critical error (e.g., API down)
		if uc.alertService != nil && isCriticalAPIError(err) {
			uc.alertService.SendCritical(ctx,
				"Stock API Critical Error",
				fmt.Sprintf("Failed to fetch price for %s: API may be down", stockCode),
				err)
		}
		return err
	}

	if err := uc.stockRepo.SaveStockPrice(ctx, price); err != nil {
		// Database errors are critical
		if uc.alertService != nil {
			uc.alertService.SendCritical(ctx,
				"Database Error",
				fmt.Sprintf("Failed to save price for %s", stockCode),
				err)
		}
		return err
	}

	logrus.Debugf("Price updated for %s: %.2f", stockCode, price.ClosePrice)
	return nil
}

// isCriticalAPIError checks if an API error is critical
func isCriticalAPIError(err error) bool {
	// Check for specific critical error types
	errorStr := err.Error()
	criticalPatterns := []string{
		"connection refused",
		"timeout",
		"rate limit",
		"unauthorized",
		"forbidden",
	}

	for _, pattern := range criticalPatterns {
		if containsIgnoreCase(errorStr, pattern) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			result[i] = c + 'a' - 'A'
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
