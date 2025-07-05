package api

import (
	"context"

	"github.com/boost-jp/stock-automation/app/usecase"
)

// DataCollector is a thin wrapper around CollectDataUseCase for backward compatibility.
type DataCollector struct {
	useCase *usecase.CollectDataUseCase
}

// NewDataCollector creates a new data collector.
func NewDataCollector(useCase *usecase.CollectDataUseCase) *DataCollector {
	return &DataCollector{
		useCase: useCase,
	}
}

// UpdateWatchList updates the watch list.
func (dc *DataCollector) UpdateWatchList() error {
	return dc.useCase.UpdateWatchList(context.Background())
}

// UpdatePortfolio updates the portfolio.
func (dc *DataCollector) UpdatePortfolio() error {
	return dc.useCase.UpdatePortfolio(context.Background())
}

// UpdateAllPrices updates prices for all stocks.
func (dc *DataCollector) UpdateAllPrices() error {
	return dc.useCase.UpdateAllPrices(context.Background())
}

// CollectHistoricalData collects historical data.
func (dc *DataCollector) CollectHistoricalData(stockCode string, days int) error {
	return dc.useCase.CollectHistoricalData(context.Background(), stockCode, days)
}

// IsMarketOpen checks if the market is open.
func (dc *DataCollector) IsMarketOpen() bool {
	return dc.useCase.IsMarketOpen()
}
