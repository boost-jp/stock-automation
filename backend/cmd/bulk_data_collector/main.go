package main

import (
	"context"
	"fmt"
	"log"
	"stock-automation/internal/api"
	"stock-automation/internal/database"
	"stock-automation/internal/models"
	"sync"
	"time"

	"gorm.io/gorm"
)

// BulkDataCollector handles bulk historical data collection for technical analysis
type BulkDataCollector struct {
	db          *gorm.DB
	yahooClient *api.YahooFinanceClient
	maxRetries  int
	maxWorkers  int
}

// NewBulkDataCollector creates a new bulk data collector
func NewBulkDataCollector(db *gorm.DB) *BulkDataCollector {
	return &BulkDataCollector{
		db:          db,
		yahooClient: api.NewYahooFinanceClient(),
		maxRetries:  3,
		maxWorkers:  3, // 並列度を3に制限（API制限を考慮）
	}
}

// CollectHistoricalData collects historical data for multiple stocks
func (bdc *BulkDataCollector) CollectHistoricalData(ctx context.Context, stockCodes []string, days int) error {
	startDate := time.Now().AddDate(0, 0, -days)

	log.Printf("📊 開始: %d銘柄の過去%d日分のデータを一括取得", len(stockCodes), days)

	for i, code := range stockCodes {
		log.Printf("📈 処理中 [%d/%d]: %s", i+1, len(stockCodes), code)

		// Check if we already have recent data for this stock
		var latestRecord models.StockPrice
		err := bdc.db.Where("code = ?", code).Order("timestamp DESC").First(&latestRecord).Error

		if err == nil && latestRecord.Timestamp.After(startDate) {
			log.Printf("✅ %s: 既存データあり (最新: %s)", code, latestRecord.Timestamp.Format("2006-01-02"))
			continue
		}

		// Collect historical data for this stock
		err = bdc.collectHistoricalDataForStock(ctx, code, startDate, time.Now())
		if err != nil {
			log.Printf("❌ %s: データ取得エラー: %v", code, err)
			continue
		}

		log.Printf("✅ %s: データ取得完了", code)

		// Rate limiting to avoid API throttling
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 2): // 2秒間隔
			// Continue to next stock
		}
	}

	log.Printf("🎉 完了: 全%d銘柄のデータ一括取得が完了しました", len(stockCodes))
	return nil
}

// CollectHistoricalDataParallel collects historical data for multiple stocks using parallel processing
func (bdc *BulkDataCollector) CollectHistoricalDataParallel(ctx context.Context, stockCodes []string, days int) error {
	startDate := time.Now().AddDate(0, 0, -days)

	log.Printf("🚀 開始: %d銘柄の過去%d日分のデータを並列取得 (最大%d並列)", len(stockCodes), days, bdc.maxWorkers)

	// Create a work channel
	jobs := make(chan string, len(stockCodes))
	results := make(chan error, len(stockCodes))

	// Create a semaphore to limit concurrent workers
	semaphore := make(chan struct{}, bdc.maxWorkers)

	var wg sync.WaitGroup

	// Start workers
	for _, code := range stockCodes {
		wg.Add(1)
		jobs <- code

		go func(stockCode string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("📈 処理開始: %s", stockCode)

			// Check if we already have recent data for this stock
			var latestRecord models.StockPrice
			err := bdc.db.Where("code = ?", stockCode).Order("timestamp DESC").First(&latestRecord).Error

			if err == nil && latestRecord.Timestamp.After(startDate) {
				log.Printf("✅ %s: 既存データあり (最新: %s)", stockCode, latestRecord.Timestamp.Format("2006-01-02"))
				results <- nil
				return
			}

			// Collect historical data with retry logic
			err = bdc.collectHistoricalDataForStockWithRetry(ctx, stockCode, startDate, time.Now())
			if err != nil {
				log.Printf("❌ %s: データ取得エラー: %v", stockCode, err)
				results <- err
				return
			}

			log.Printf("✅ %s: データ取得完了", stockCode)
			results <- nil
		}(code)
	}

	// Close jobs channel
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Collect results
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		log.Printf("⚠️  %d件のエラーが発生しました", len(errors))
		for _, err := range errors {
			log.Printf("   - %v", err)
		}
	}

	log.Printf("🎉 完了: 全%d銘柄のデータ並列取得が完了しました (エラー: %d件)", len(stockCodes), len(errors))
	return nil
}

// collectHistoricalDataForStockWithRetry collects data with retry logic
func (bdc *BulkDataCollector) collectHistoricalDataForStockWithRetry(ctx context.Context, code string, startDate, endDate time.Time) error {
	var lastErr error

	for attempt := 1; attempt <= bdc.maxRetries; attempt++ {
		err := bdc.collectHistoricalDataForStock(ctx, code, startDate, endDate)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < bdc.maxRetries {
			waitTime := time.Duration(attempt) * time.Second * 2 // Exponential backoff
			log.Printf("🔄 %s: リトライ中 (%d/%d) - %v秒後に再試行", code, attempt, bdc.maxRetries, waitTime.Seconds())

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// Continue to next attempt
			}
		}
	}

	return fmt.Errorf("最大リトライ回数(%d)に達しました: %w", bdc.maxRetries, lastErr)
}

// collectHistoricalDataForStock collects historical data for a single stock using Yahoo Finance API
func (bdc *BulkDataCollector) collectHistoricalDataForStock(ctx context.Context, code string, startDate, endDate time.Time) error {
	// Calculate number of days to fetch
	days := int(endDate.Sub(startDate).Hours() / 24)

	// Use Yahoo Finance API to get historical data
	stockPrices, err := bdc.yahooClient.GetHistoricalData(code, days)
	if err != nil {
		return fmt.Errorf("failed to fetch historical data from Yahoo Finance: %w", err)
	}

	// Get stock name from the first record or set a default
	stockName := fmt.Sprintf("Stock_%s", code)
	if len(stockPrices) > 0 {
		// You might want to fetch the actual stock name from another API or maintain a mapping
		stockName = bdc.getStockName(code)
	}

	// Set the stock name for all records and validate
	validPrices := make([]models.StockPrice, 0)
	for i := range stockPrices {
		stockPrices[i].Name = stockName
		// Validate the data
		if stockPrices[i].IsValid() {
			validPrices = append(validPrices, stockPrices[i])
		} else {
			log.Printf("⚠️  無効なデータをスキップ: %s at %s", code, stockPrices[i].Timestamp.Format("2006-01-02"))
		}
	}

	// Batch upsert to database (insert or update on conflict)
	if len(validPrices) > 0 {
		err := bdc.upsertStockPrices(validPrices)
		if err != nil {
			return fmt.Errorf("failed to batch upsert stock prices: %w", err)
		}
		log.Printf("💾 %s: %d件のデータを保存/更新しました", code, len(validPrices))
	}

	return nil
}

// upsertStockPrices performs batch upsert operation for stock prices
func (bdc *BulkDataCollector) upsertStockPrices(stockPrices []models.StockPrice) error {
	// MySQL用のON DUPLICATE KEY UPDATE構文を使用
	// 重複した場合（code + timestampの組み合わせ）は値を更新

	batchSize := 100
	for i := 0; i < len(stockPrices); i += batchSize {
		end := i + batchSize
		if end > len(stockPrices) {
			end = len(stockPrices)
		}

		batch := stockPrices[i:end]

		// Use Clauses for UPSERT operation
		err := bdc.db.Clauses().CreateInBatches(batch, batchSize).Error
		if err != nil {
			// Fallback to individual upserts if batch fails
			for _, price := range batch {
				err = bdc.db.Where("code = ? AND DATE(timestamp) = DATE(?)",
					price.Code, price.Timestamp).
					Assign(price).
					FirstOrCreate(&price).Error
				if err != nil {
					return fmt.Errorf("failed to upsert stock price for %s: %w", price.Code, err)
				}
			}
		}
	}

	return nil
}

// getStockName returns the stock name for a given code
func (bdc *BulkDataCollector) getStockName(code string) string {
	// Mapping of stock codes to names for major Japanese stocks
	stockNames := map[string]string{
		"7203": "トヨタ自動車",
		"6758": "ソニーグループ",
		"9984": "ソフトバンクグループ",
		"8306": "三菱UFJフィナンシャル・グループ",
		"6861": "キーエンス",
		"4063": "信越化学工業",
		"6954": "ファナック",
		"9432": "日本電信電話",
		"4523": "エーザイ",
		"6501": "日立製作所",
	}

	if name, exists := stockNames[code]; exists {
		return name
	}
	return fmt.Sprintf("Stock_%s", code)
}

// GetStockCodesForAnalysis returns the list of stock codes to analyze
func (bdc *BulkDataCollector) GetStockCodesForAnalysis() []string {
	// These would come from a configuration file or database
	// For now, returning some major Japanese stocks
	return []string{
		"7203", // Toyota
		"6758", // Sony
		"9984", // SoftBank
		"8306", // Mitsubishi UFJ
		"6861", // Keyence
		"4063", // Shin-Etsu Chemical
		"6954", // Fanuc
		"9432", // NTT
		"4523", // Eisai
		"6501", // Hitachi
	}
}

func main() {
	// Initialize database connection
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
	}

	// Create bulk data collector
	bulkCollector := NewBulkDataCollector(db.GetDB())

	// Get stock codes for analysis
	stockCodes := bulkCollector.GetStockCodesForAnalysis()

	// Collect historical data for the past 365 days using parallel processing
	ctx := context.Background()
	err = bulkCollector.CollectHistoricalDataParallel(ctx, stockCodes, 365)
	if err != nil {
		log.Fatal("データ一括取得エラー:", err)
	}

	log.Println("📊 テクニカル分析用データの並列一括取得が完了しました")
}
