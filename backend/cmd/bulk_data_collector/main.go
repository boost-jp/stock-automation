package main

import (
	"context"
	"fmt"
	"log"
	"stock-automation/internal/api"
	"stock-automation/internal/database"
	"stock-automation/internal/models"
	"time"

	"gorm.io/gorm"
)

// BulkDataCollector handles bulk historical data collection for technical analysis
type BulkDataCollector struct {
	db          *gorm.DB
	yahooClient *api.YahooFinanceClient
}

// NewBulkDataCollector creates a new bulk data collector
func NewBulkDataCollector(db *gorm.DB) *BulkDataCollector {
	return &BulkDataCollector{
		db:          db,
		yahooClient: api.NewYahooFinanceClient(),
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

	// Batch insert to database
	if len(validPrices) > 0 {
		err := bdc.db.CreateInBatches(validPrices, 100).Error
		if err != nil {
			return fmt.Errorf("failed to batch insert stock prices: %w", err)
		}
		log.Printf("💾 %s: %d件のデータを保存しました", code, len(validPrices))
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

	// Collect historical data for the past 365 days
	ctx := context.Background()
	err = bulkCollector.CollectHistoricalData(ctx, stockCodes, 365)
	if err != nil {
		log.Fatal("データ一括取得エラー:", err)
	}

	log.Println("📊 テクニカル分析用データの一括取得が完了しました")
}
