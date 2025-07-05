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

// StockPrice validation methods
func (s *StockPrice) IsValid() bool {
	if s.Code == "" || s.Name == "" {
		return false
	}
	if s.Price < 0 || s.High < 0 || s.Low < 0 || s.Open < 0 || s.Close < 0 {
		return false
	}
	if s.Volume < 0 {
		return false
	}
	return true
}

func (s *StockPrice) CalculateGainLoss(portfolio Portfolio) float64 {
	return s.Price - portfolio.PurchasePrice
}

func (s *StockPrice) CalculateReturnRate(portfolio Portfolio) float64 {
	if portfolio.PurchasePrice == 0 {
		return 0
	}
	return (s.Price - portfolio.PurchasePrice) / portfolio.PurchasePrice * 100
}

// TechnicalIndicator validation methods
func (t *TechnicalIndicator) IsValid() bool {
	if t.Code == "" {
		return false
	}
	if t.RSI < 0 || t.RSI > 100 {
		return false
	}
	return true
}

func (t *TechnicalIndicator) GetSignalStrength() string {
	buySignals := 0
	sellSignals := 0

	// RSI signals
	if t.RSI < 30 {
		buySignals++
	} else if t.RSI > 70 {
		sellSignals++
	}

	// MACD signals
	if t.MACD > t.Signal && t.Histogram > 0 {
		buySignals++
	} else if t.MACD < t.Signal && t.Histogram < 0 {
		sellSignals++
	}

	if buySignals > sellSignals {
		return "Strong Buy"
	} else if sellSignals > buySignals {
		return "Strong Sell"
	}
	return "Neutral"
}

// Portfolio validation methods
func (p *Portfolio) IsValid() bool {
	if p.Code == "" || p.Name == "" {
		return false
	}
	if p.Shares <= 0 {
		return false
	}
	if p.PurchasePrice <= 0 {
		return false
	}
	return true
}

func (p *Portfolio) CalculateTotalValue(currentPrice float64) float64 {
	return float64(p.Shares) * currentPrice
}

func (p *Portfolio) GetPurchaseValue() float64 {
	return float64(p.Shares) * p.PurchasePrice
}

// WatchList validation methods
func (w *WatchList) IsValid() bool {
	if w.Code == "" || w.Name == "" {
		return false
	}
	if w.TargetBuyPrice > 0 && w.TargetSellPrice > 0 && w.TargetBuyPrice >= w.TargetSellPrice {
		return false
	}
	return true
}
