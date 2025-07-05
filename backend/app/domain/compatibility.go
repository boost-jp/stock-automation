package domain

import (
	"github.com/boost-jp/stock-automation/app/domain/models"
)

// Package-level service instances for compatibility
var (
	portfolioService         = NewPortfolioService()
	technicalAnalysisService = NewTechnicalAnalysisService()
)

// CalculatePortfolioSummary is a compatibility wrapper for the portfolio service method
func CalculatePortfolioSummary(portfolio []*models.Portfolio, currentPrices map[string]float64) *PortfolioSummary {
	return portfolioService.CalculatePortfolioSummary(portfolio, currentPrices)
}

// GeneratePortfolioReport is a compatibility wrapper for the portfolio service method
func GeneratePortfolioReport(summary *PortfolioSummary) string {
	return portfolioService.GeneratePortfolioReport(summary)
}

// CalculateAllIndicators is a compatibility wrapper for the technical analysis service method
func CalculateAllIndicators(prices []models.StockPrice) *models.TechnicalIndicator {
	// Convert to StockPriceData
	priceData := make([]StockPriceData, len(prices))
	for i, p := range prices {
		priceData[i] = StockPriceData{
			Code:      p.Code,
			Date:      p.Date,
			Open:      technicalAnalysisService.decimalToFloat(p.OpenPrice),
			High:      technicalAnalysisService.decimalToFloat(p.HighPrice),
			Low:       technicalAnalysisService.decimalToFloat(p.LowPrice),
			Close:     technicalAnalysisService.decimalToFloat(p.ClosePrice),
			Volume:    p.Volume,
			Timestamp: p.Date,
		}
	}

	// Calculate indicators
	indicatorData := technicalAnalysisService.CalculateAllIndicators(priceData)
	if indicatorData == nil {
		return nil
	}

	// Convert back to model
	return technicalAnalysisService.ConvertToModelIndicator(indicatorData)
}
