package usecase

import (
	"context"
	"fmt"

	"github.com/boost-jp/stock-automation/app/domain"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/sirupsen/logrus"
)

// TechnicalAnalysisUseCase handles technical analysis business logic.
type TechnicalAnalysisUseCase struct {
	stockRepo   repository.StockRepository
	stockClient client.StockDataClient
}

// NewTechnicalAnalysisUseCase creates a new technical analysis use case.
func NewTechnicalAnalysisUseCase(
	stockRepo repository.StockRepository,
	stockClient client.StockDataClient,
) *TechnicalAnalysisUseCase {
	return &TechnicalAnalysisUseCase{
		stockRepo:   stockRepo,
		stockClient: stockClient,
	}
}

// CalculateAndSaveTechnicalIndicators calculates and saves technical indicators for a stock.
func (uc *TechnicalAnalysisUseCase) CalculateAndSaveTechnicalIndicators(ctx context.Context, stockCode string) error {
	// Get historical prices
	prices, err := uc.stockRepo.GetPriceHistory(ctx, stockCode, 100)
	if err != nil {
		return fmt.Errorf("failed to get price history: %w", err)
	}

	if len(prices) < 20 {
		return fmt.Errorf("insufficient data for technical analysis: %d records", len(prices))
	}

	// Convert to non-pointer slice for analysis functions
	priceValues := make([]models.StockPrice, len(prices))
	for i, p := range prices {
		priceValues[i] = *p
	}

	// Use the existing analysis functions
	indicator := domain.CalculateAllIndicators(priceValues)

	// Set the stock code (indicator already has the correct structure)
	if indicator == nil {
		return fmt.Errorf("failed to calculate indicators")
	}
	indicator.Code = stockCode

	// Save to database
	if err := uc.stockRepo.SaveTechnicalIndicator(ctx, indicator); err != nil {
		return fmt.Errorf("failed to save technical indicator: %w", err)
	}

	logrus.Infof("Technical indicators calculated and saved for %s", stockCode)
	return nil
}

// GetTechnicalAnalysis retrieves the latest technical analysis for a stock.
func (uc *TechnicalAnalysisUseCase) GetTechnicalAnalysis(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error) {
	indicator, err := uc.stockRepo.GetLatestTechnicalIndicator(ctx, stockCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get technical indicator: %w", err)
	}
	return indicator, nil
}

// AnalyzeWatchList performs technical analysis on all watched stocks.
func (uc *TechnicalAnalysisUseCase) AnalyzeWatchList(ctx context.Context) error {
	// Get active watch list
	watchList, err := uc.stockRepo.GetActiveWatchList(ctx)
	if err != nil {
		return fmt.Errorf("failed to get watch list: %w", err)
	}

	// Analyze each stock
	for _, item := range watchList {
		if err := uc.CalculateAndSaveTechnicalIndicators(ctx, item.Code); err != nil {
			logrus.Errorf("Failed to analyze %s: %v", item.Code, err)
			continue
		}
	}

	logrus.Infof("Technical analysis completed for %d stocks", len(watchList))
	return nil
}

// GetTradingSignals generates trading signals based on technical indicators.
func (uc *TechnicalAnalysisUseCase) GetTradingSignals(ctx context.Context, stockCode string) ([]string, error) {
	indicator, err := uc.stockRepo.GetLatestTechnicalIndicator(ctx, stockCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get technical indicator: %w", err)
	}

	var signals []string

	// RSI signals
	rsi := client.NullDecimalToFloat(indicator.Rsi14)
	if rsi > 0 {
		if rsi < 30 {
			signals = append(signals, "RSI買いシグナル（売られすぎ）")
		} else if rsi > 70 {
			signals = append(signals, "RSI売りシグナル（買われすぎ）")
		}
	}

	// MACD signals
	macd := client.NullDecimalToFloat(indicator.Macd)
	signal := client.NullDecimalToFloat(indicator.MacdSignal)
	if macd > 0 && signal > 0 {
		if macd > signal {
			signals = append(signals, "MACDゴールデンクロス（買いシグナル）")
		} else if macd < signal {
			signals = append(signals, "MACDデッドクロス（売りシグナル）")
		}
	}

	// Moving average signals based on current price position
	currentPrice, err := uc.stockRepo.GetLatestPrice(ctx, stockCode)
	if err == nil {
		price := client.DecimalToFloat(currentPrice.ClosePrice)
		sma5 := client.NullDecimalToFloat(indicator.Sma5)
		sma25 := client.NullDecimalToFloat(indicator.Sma25)

		if sma5 > 0 && sma25 > 0 {
			if price > sma5 && sma5 > sma25 {
				signals = append(signals, "上昇トレンド（価格 > 5日移動平均 > 25日移動平均）")
			} else if price < sma5 && sma5 < sma25 {
				signals = append(signals, "下降トレンド（価格 < 5日移動平均 < 25日移動平均）")
			}
		}
	}

	return signals, nil
}
