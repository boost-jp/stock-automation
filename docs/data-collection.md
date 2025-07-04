# 株価データ収集システム

## 概要
Yahoo Finance APIを使用したリアルタイム株価データの取得・保存システム

## データソース

### Yahoo Finance API
- **BASE URL**: `https://query1.finance.yahoo.com`
- **利点**: 無料、高頻度アクセス可能、豊富なデータ
- **制限**: 非公式API、仕様変更の可能性

### 取得可能データ
- 現在価格・出来高
- 日足・分足データ（最大2年分）
- 財務指標（PER、PBR等）
- 企業情報・配当情報

## Go実装

### データ構造定義

#### `internal/models/stock.go`
```go
package models

import (
    "time"
)

// 株価データ
type StockPrice struct {
    ID        uint      `gorm:"primaryKey"`
    Code      string    `gorm:"index;not null"`
    Name      string    `gorm:"not null"`
    Price     float64   `gorm:"not null"`
    Volume    int64     `gorm:"not null"`
    High      float64   `gorm:"not null"`
    Low       float64   `gorm:"not null"`
    Open      float64   `gorm:"not null"`
    Close     float64   `gorm:"not null"`
    Timestamp time.Time `gorm:"index;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

// テクニカル指標
type TechnicalIndicator struct {
    ID        uint      `gorm:"primaryKey"`
    Code      string    `gorm:"index;not null"`
    MA5       float64   `gorm:"column:ma5"`
    MA25      float64   `gorm:"column:ma25"`
    MA75      float64   `gorm:"column:ma75"`
    RSI       float64   `gorm:"column:rsi"`
    MACD      float64   `gorm:"column:macd"`
    Signal    float64   `gorm:"column:signal"`
    Histogram float64   `gorm:"column:histogram"`
    Timestamp time.Time `gorm:"index;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

// ポートフォリオ
type Portfolio struct {
    ID            uint      `gorm:"primaryKey"`
    Code          string    `gorm:"index;not null"`
    Name          string    `gorm:"not null"`
    Shares        int       `gorm:"not null"`
    PurchasePrice float64   `gorm:"not null"`
    PurchaseDate  time.Time `gorm:"not null"`
    CreatedAt     time.Time `gorm:"autoCreateTime"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// 監視銘柄
type WatchList struct {
    ID              uint      `gorm:"primaryKey"`
    Code            string    `gorm:"uniqueIndex;not null"`
    Name            string    `gorm:"not null"`
    TargetBuyPrice  float64   `gorm:"column:target_buy_price"`
    TargetSellPrice float64   `gorm:"column:target_sell_price"`
    IsActive        bool      `gorm:"default:true"`
    CreatedAt       time.Time `gorm:"autoCreateTime"`
    UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}
```

### Yahoo Finance APIクライント

#### `internal/api/yahoo_finance.go`
```go
package api

import (
    "encoding/json"
    "fmt"
    "strconv"
    "time"
    
    "github.com/go-resty/resty/v2"
    "github.com/sirupsen/logrus"
)

type YahooFinanceClient struct {
    client  *resty.Client
    baseURL string
}

// Yahoo Finance APIレスポンス構造
type YahooFinanceResponse struct {
    Chart struct {
        Result []struct {
            Meta struct {
                Symbol                string  `json:"symbol"`
                RegularMarketPrice    float64 `json:"regularMarketPrice"`
                PreviousClose         float64 `json:"previousClose"`
                RegularMarketOpen     float64 `json:"regularMarketOpen"`
                RegularMarketDayLow   float64 `json:"regularMarketDayLow"`
                RegularMarketDayHigh  float64 `json:"regularMarketDayHigh"`
                RegularMarketVolume   int64   `json:"regularMarketVolume"`
                Currency              string  `json:"currency"`
                ExchangeName          string  `json:"exchangeName"`
            } `json:"meta"`
            Timestamp []int64 `json:"timestamp"`
            Indicators struct {
                Quote []struct {
                    Open   []float64 `json:"open"`
                    High   []float64 `json:"high"`
                    Low    []float64 `json:"low"`
                    Close  []float64 `json:"close"`
                    Volume []int64   `json:"volume"`
                } `json:"quote"`
            } `json:"indicators"`
        } `json:"result"`
        Error interface{} `json:"error"`
    } `json:"chart"`
}

func NewYahooFinanceClient() *YahooFinanceClient {
    client := resty.New()
    client.SetTimeout(30 * time.Second)
    client.SetRetryCount(3)
    client.SetRetryWaitTime(1 * time.Second)
    
    return &YahooFinanceClient{
        client:  client,
        baseURL: "https://query1.finance.yahoo.com",
    }
}

// リアルタイム株価取得
func (y *YahooFinanceClient) GetCurrentPrice(stockCode string) (*StockPrice, error) {
    url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)
    
    resp, err := y.client.R().
        SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
        Get(url)
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch data: %w", err)
    }
    
    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode())
    }
    
    var response YahooFinanceResponse
    if err := json.Unmarshal(resp.Body(), &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    if len(response.Chart.Result) == 0 {
        return nil, fmt.Errorf("no data found for stock code: %s", stockCode)
    }
    
    result := response.Chart.Result[0]
    meta := result.Meta
    
    stockPrice := &StockPrice{
        Code:      stockCode,
        Price:     meta.RegularMarketPrice,
        Volume:    meta.RegularMarketVolume,
        High:      meta.RegularMarketDayHigh,
        Low:       meta.RegularMarketDayLow,
        Open:      meta.RegularMarketOpen,
        Close:     meta.RegularMarketPrice, // 現在価格を終値として使用
        Timestamp: time.Now(),
    }
    
    logrus.Debugf("Fetched data for %s: ¥%.2f", stockCode, stockPrice.Price)
    return stockPrice, nil
}

// 履歴データ取得
func (y *YahooFinanceClient) GetHistoricalData(stockCode string, days int) ([]StockPrice, error) {
    endTime := time.Now().Unix()
    startTime := time.Now().AddDate(0, 0, -days).Unix()
    
    url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)
    
    resp, err := y.client.R().
        SetQueryParams(map[string]string{
            "period1":  strconv.FormatInt(startTime, 10),
            "period2":  strconv.FormatInt(endTime, 10),
            "interval": "1d",
        }).
        SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
        Get(url)
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch historical data: %w", err)
    }
    
    var response YahooFinanceResponse
    if err := json.Unmarshal(resp.Body(), &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    if len(response.Chart.Result) == 0 {
        return nil, fmt.Errorf("no historical data found for: %s", stockCode)
    }
    
    result := response.Chart.Result[0]
    timestamps := result.Timestamp
    quotes := result.Indicators.Quote[0]
    
    var prices []StockPrice
    for i, ts := range timestamps {
        if i >= len(quotes.Close) || quotes.Close[i] == 0 {
            continue
        }
        
        price := StockPrice{
            Code:      stockCode,
            Open:      quotes.Open[i],
            High:      quotes.High[i],
            Low:       quotes.Low[i],
            Close:     quotes.Close[i],
            Price:     quotes.Close[i],
            Volume:    quotes.Volume[i],
            Timestamp: time.Unix(ts, 0),
        }
        
        prices = append(prices, price)
    }
    
    logrus.Debugf("Fetched %d historical records for %s", len(prices), stockCode)
    return prices, nil
}

// 分足データ取得
func (y *YahooFinanceClient) GetIntradayData(stockCode string, interval string) ([]StockPrice, error) {
    // interval: 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo
    url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)
    
    resp, err := y.client.R().
        SetQueryParams(map[string]string{
            "range":    "1d",
            "interval": interval,
        }).
        SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
        Get(url)
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch intraday data: %w", err)
    }
    
    var response YahooFinanceResponse
    if err := json.Unmarshal(resp.Body(), &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    if len(response.Chart.Result) == 0 {
        return nil, fmt.Errorf("no intraday data found for: %s", stockCode)
    }
    
    result := response.Chart.Result[0]
    timestamps := result.Timestamp
    quotes := result.Indicators.Quote[0]
    
    var prices []StockPrice
    for i, ts := range timestamps {
        if i >= len(quotes.Close) || quotes.Close[i] == 0 {
            continue
        }
        
        price := StockPrice{
            Code:      stockCode,
            Open:      quotes.Open[i],
            High:      quotes.High[i],
            Low:       quotes.Low[i],
            Close:     quotes.Close[i],
            Price:     quotes.Close[i],
            Volume:    quotes.Volume[i],
            Timestamp: time.Unix(ts, 0),
        }
        
        prices = append(prices, price)
    }
    
    return prices, nil
}
```

### データ収集サービス

#### `internal/api/data_collector.go`
```go
package api

import (
    "sync"
    "time"
    
    "stock-automation/internal/database"
    "stock-automation/internal/models"
    
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

// 監視銘柄リストの更新
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

// ポートフォリオの更新
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

// 全銘柄の価格更新
func (dc *DataCollector) UpdateAllPrices() error {
    dc.mutex.RLock()
    allStocks := make(map[string]string) // code -> name
    
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
    semaphore := make(chan struct{}, 5) // 同時実行数制限
    
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

// 単一銘柄の価格更新
func (dc *DataCollector) updateSinglePrice(stockCode, stockName string) error {
    price, err := dc.yahooClient.GetCurrentPrice(stockCode)
    if err != nil {
        return err
    }
    
    price.Name = stockName
    
    // データベースに保存
    if err := dc.db.SaveStockPrice(price); err != nil {
        return err
    }
    
    logrus.Debugf("Updated %s (%s): ¥%.2f", stockName, stockCode, price.Price)
    return nil
}

// 履歴データの一括取得
func (dc *DataCollector) CollectHistoricalData(stockCode string, days int) error {
    prices, err := dc.yahooClient.GetHistoricalData(stockCode, days)
    if err != nil {
        return err
    }
    
    // 銘柄名の取得
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
    
    // 銘柄名設定
    for i := range prices {
        prices[i].Name = stockName
    }
    
    // 一括保存
    if err := dc.db.SaveStockPrices(prices); err != nil {
        return err
    }
    
    logrus.Infof("Collected %d historical records for %s", len(prices), stockCode)
    return nil
}

// 市場時間チェック
func (dc *DataCollector) IsMarketOpen() bool {
    now := time.Now()
    
    // 土日は休場
    weekday := now.Weekday()
    if weekday == time.Saturday || weekday == time.Sunday {
        return false
    }
    
    // 平日9:00-15:00（JST）
    hour := now.Hour()
    return hour >= 9 && hour < 15
}

// データ品質チェック
func (dc *DataCollector) ValidateData(price *models.StockPrice) bool {
    // 基本的なバリデーション
    if price.Price <= 0 || price.Volume < 0 {
        return false
    }
    
    if price.High < price.Low {
        return false
    }
    
    if price.Price > price.High*1.1 || price.Price < price.Low*0.9 {
        return false
    }
    
    return true
}
```

### データベース操作

#### `internal/database/stock_operations.go`
```go
package database

import (
    "time"
    
    "stock-automation/internal/models"
    
    "gorm.io/gorm"
)

// 株価データ保存
func (db *DB) SaveStockPrice(price *models.StockPrice) error {
    return db.conn.Create(price).Error
}

// 株価データ一括保存
func (db *DB) SaveStockPrices(prices []models.StockPrice) error {
    return db.conn.CreateInBatches(prices, 100).Error
}

// 最新株価取得
func (db *DB) GetLatestPrice(stockCode string) (*models.StockPrice, error) {
    var price models.StockPrice
    err := db.conn.Where("code = ?", stockCode).
        Order("timestamp desc").
        First(&price).Error
    
    if err != nil {
        return nil, err
    }
    
    return &price, nil
}

// 期間内の株価データ取得
func (db *DB) GetPriceHistory(stockCode string, days int) ([]models.StockPrice, error) {
    var prices []models.StockPrice
    startTime := time.Now().AddDate(0, 0, -days)
    
    err := db.conn.Where("code = ? AND timestamp >= ?", stockCode, startTime).
        Order("timestamp asc").
        Find(&prices).Error
    
    return prices, err
}

// 監視銘柄リスト取得
func (db *DB) GetActiveWatchList() ([]models.WatchList, error) {
    var watchList []models.WatchList
    err := db.conn.Where("is_active = ?", true).Find(&watchList).Error
    return watchList, err
}

// ポートフォリオ取得
func (db *DB) GetPortfolio() ([]models.Portfolio, error) {
    var portfolio []models.Portfolio
    err := db.conn.Find(&portfolio).Error
    return portfolio, err
}

// 古いデータの削除（データベース容量管理）
func (db *DB) CleanupOldData(days int) error {
    cutoffTime := time.Now().AddDate(0, 0, -days)
    
    return db.conn.Where("timestamp < ?", cutoffTime).
        Delete(&models.StockPrice{}).Error
}
```

### スケジューラー統合

#### `internal/api/scheduler.go`
```go
package api

import (
    "time"
    
    "github.com/go-co-op/gocron"
    "github.com/sirupsen/logrus"
)

type DataScheduler struct {
    collector *DataCollector
    scheduler *gocron.Scheduler
}

func NewDataScheduler(collector *DataCollector) *DataScheduler {
    s := gocron.NewScheduler(time.UTC)
    
    return &DataScheduler{
        collector: collector,
        scheduler: s,
    }
}

func (ds *DataScheduler) StartScheduledCollection() {
    // 5分毎の価格更新（市場時間中のみ）
    ds.scheduler.Every(5).Minutes().Do(func() {
        if ds.collector.IsMarketOpen() {
            if err := ds.collector.UpdateAllPrices(); err != nil {
                logrus.Error("Failed to update prices:", err)
            }
        }
    })
    
    // 30分毎の設定更新
    ds.scheduler.Every(30).Minutes().Do(func() {
        if err := ds.collector.UpdateWatchList(); err != nil {
            logrus.Error("Failed to update watch list:", err)
        }
        if err := ds.collector.UpdatePortfolio(); err != nil {
            logrus.Error("Failed to update portfolio:", err)
        }
    })
    
    // 毎日深夜のデータクリーンアップ
    ds.scheduler.Every(1).Day().At("02:00").Do(func() {
        if err := ds.collector.db.CleanupOldData(365); err != nil {
            logrus.Error("Failed to cleanup old data:", err)
        }
    })
    
    ds.scheduler.StartAsync()
    logrus.Info("Data collection scheduler started")
}

func (ds *DataScheduler) Stop() {
    ds.scheduler.Stop()
    logrus.Info("Data collection scheduler stopped")
}
```

### エラーハンドリング・リトライ

```go
func (dc *DataCollector) updateSinglePriceWithRetry(stockCode, stockName string, maxRetries int) error {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        price, err := dc.yahooClient.GetCurrentPrice(stockCode)
        if err != nil {
            lastErr = err
            logrus.Warnf("Attempt %d/%d failed for %s: %v", attempt, maxRetries, stockCode, err)
            
            if attempt < maxRetries {
                time.Sleep(time.Duration(attempt) * time.Second)
                continue
            }
            return lastErr
        }
        
        price.Name = stockName
        
        if !dc.ValidateData(price) {
            lastErr = fmt.Errorf("invalid data for %s", stockCode)
            continue
        }
        
        if err := dc.db.SaveStockPrice(price); err != nil {
            lastErr = err
            continue
        }
        
        logrus.Debugf("Updated %s (%s): ¥%.2f (attempt %d)", stockName, stockCode, price.Price, attempt)
        return nil
    }
    
    return lastErr
}
```

この実装により、効率的で信頼性の高い株価データ収集システムが構築できます。次は[判断アルゴリズム](decision-algorithm.md)の実装に進んでください。