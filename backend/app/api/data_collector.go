package api

import (
	"sync"
	"time"

	"github.com/boost-jp/stock-automation/app/database"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/sirupsen/logrus"
)

type DataCollector struct {
	yahooClient *YahooFinanceClient
	db          *database.DB
	watchList   []models.WatchList
	portfolio   []models.Portfolio
	mutex       sync.RWMutex
}

func NewDataCollector(db *database.DB) *DataCollector {
	return &DataCollector{
		yahooClient: NewYahooFinanceClient(),
		db:          db,
	}
}

// 監視銘柄リストの更新.
func (dc *DataCollector) UpdateWatchList() error {
	watchList, err := dc.db.GetActiveWatchList()
	if err != nil {
		return err
	}

	dc.mutex.Lock()
	dc.watchList = watchList
	dc.mutex.Unlock()

	logrus.Infof("Updated watch list: %d stocks", len(watchList))

	return nil
}

// ポートフォリオの更新.
func (dc *DataCollector) UpdatePortfolio() error {
	portfolio, err := dc.db.GetPortfolio()
	if err != nil {
		return err
	}

	dc.mutex.Lock()
	dc.portfolio = portfolio
	dc.mutex.Unlock()

	logrus.Infof("Updated portfolio: %d stocks", len(portfolio))

	return nil
}

// 全銘柄の価格更新.
func (dc *DataCollector) UpdateAllPrices() error {
	dc.mutex.RLock()
	allStocks := make(map[string]string)

	// 監視銘柄
	for _, stock := range dc.watchList {
		allStocks[stock.Code] = stock.Name
	}

	// ポートフォリオ銘柄
	for _, stock := range dc.portfolio {
		allStocks[stock.Code] = stock.Name
	}
	dc.mutex.RUnlock()

	if len(allStocks) == 0 {
		logrus.Debug("No stocks to update")

		return nil
	}

	// 並行処理で価格取得
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 5)

	for code, name := range allStocks {
		wg.Add(1)

		go func(stockCode, stockName string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := dc.updateSinglePrice(stockCode, stockName); err != nil {
				logrus.Errorf("Failed to update price for %s: %v", stockCode, err)
			}
		}(code, name)
	}

	wg.Wait()
	logrus.Infof("Updated prices for %d stocks", len(allStocks))

	return nil
}

// 単一銘柄の価格更新.
func (dc *DataCollector) updateSinglePrice(stockCode, stockName string) error {
	price, err := dc.yahooClient.GetCurrentPrice(stockCode)
	if err != nil {
		return err
	}

	price.Name = stockName

	if err := dc.db.SaveStockPrice(price); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"code":  stockCode,
		"name":  stockName,
		"price": client.DecimalToFloat(price.ClosePrice),
	}).Debug("Stock price updated")

	return nil
}

// 履歴データの一括取得.
func (dc *DataCollector) CollectHistoricalData(stockCode string, days int) error {
	prices, err := dc.yahooClient.GetHistoricalData(stockCode, days)
	if err != nil {
		return err
	}

	stockName := ""

	dc.mutex.RLock()
	for _, stock := range dc.watchList {
		if stock.Code == stockCode {
			stockName = stock.Name

			break
		}
	}

	if stockName == "" {
		for _, stock := range dc.portfolio {
			if stock.Code == stockCode {
				stockName = stock.Name

				break
			}
		}
	}
	dc.mutex.RUnlock()

	for i := range prices {
		prices[i].Name = stockName
	}

	if err := dc.db.SaveStockPrices(prices); err != nil {
		return err
	}

	logrus.Infof("Collected %d historical records for %s", len(prices), stockCode)

	return nil
}

// 市場時間チェック.
func (dc *DataCollector) IsMarketOpen() bool {
	now := time.Now()

	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	hour := now.Hour()

	return hour >= 9 && hour < 15
}

// データ品質チェック.
func (dc *DataCollector) ValidateData(price *models.StockPrice) bool {
	closePrice := client.DecimalToFloat(price.ClosePrice)
	if closePrice <= 0 || price.Volume < 0 {
		return false
	}

	highPrice := client.DecimalToFloat(price.HighPrice)
	lowPrice := client.DecimalToFloat(price.LowPrice)
	if highPrice < lowPrice {
		return false
	}

	if closePrice > highPrice*1.1 || closePrice < lowPrice*0.9 {
		return false
	}

	return true
}
