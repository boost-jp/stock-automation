package database

import (
	"time"

	"github.com/boost-jp/stock-automation/app/models"
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

// テクニカル指標保存
func (db *DB) SaveTechnicalIndicator(indicator *models.TechnicalIndicator) error {
	return db.conn.Create(indicator).Error
}

// 最新テクニカル指標取得
func (db *DB) GetLatestTechnicalIndicator(stockCode string) (*models.TechnicalIndicator, error) {
	var indicator models.TechnicalIndicator
	err := db.conn.Where("code = ?", stockCode).
		Order("timestamp desc").
		First(&indicator).Error

	if err != nil {
		return nil, err
	}

	return &indicator, nil
}
