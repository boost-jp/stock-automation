package models

import (
	"time"
)

// 株価データ
type StockPrice struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"index;not null;size:10"`
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
	Code      string    `gorm:"index;not null;size:10"`
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
	Code          string    `gorm:"index;not null;size:10"`
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
	Code            string    `gorm:"uniqueIndex;not null;size:10"`
	Name            string    `gorm:"not null"`
	TargetBuyPrice  float64   `gorm:"column:target_buy_price"`
	TargetSellPrice float64   `gorm:"column:target_sell_price"`
	IsActive        bool      `gorm:"default:true"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}