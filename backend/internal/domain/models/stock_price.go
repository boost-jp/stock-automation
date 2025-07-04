package models

import (
	"time"
)

// StockPrice represents a stock price data point
type StockPrice struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Date       time.Time `json:"date"`
	OpenPrice  float64   `json:"open_price"`
	HighPrice  float64   `json:"high_price"`
	LowPrice   float64   `json:"low_price"`
	ClosePrice float64   `json:"close_price"`
	Volume     int64     `json:"volume"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TechnicalIndicator represents technical indicator data
type TechnicalIndicator struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Date          time.Time `json:"date"`
	RSI14         *float64  `json:"rsi_14,omitempty"`
	MACD          *float64  `json:"macd,omitempty"`
	MACDSignal    *float64  `json:"macd_signal,omitempty"`
	MACDHistogram *float64  `json:"macd_histogram,omitempty"`
	SMA5          *float64  `json:"sma_5,omitempty"`
	SMA25         *float64  `json:"sma_25,omitempty"`
	SMA75         *float64  `json:"sma_75,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Portfolio represents a portfolio holding
type Portfolio struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Shares        int       `json:"shares"`
	PurchasePrice float64   `json:"purchase_price"`
	PurchaseDate  time.Time `json:"purchase_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WatchList represents a watch list entry
type WatchList struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	TargetBuyPrice  *float64  `json:"target_buy_price,omitempty"`
	TargetSellPrice *float64  `json:"target_sell_price,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}