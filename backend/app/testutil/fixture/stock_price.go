package fixture

import (
	"context"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
	"github.com/oklog/ulid/v2"
)

// StockPriceBuilder helps build test stock price data
type StockPriceBuilder struct {
	stockPrice *dao.StockPrice
}

// NewStockPrice creates a new stock price builder with default values
func NewStockPrice() *StockPriceBuilder {
	return &StockPriceBuilder{
		stockPrice: &dao.StockPrice{
			ID:         ulid.MustNew(ulid.Now(), nil).String(),
			Code:       "7203",
			Date:       time.Now(),
			OpenPrice:  2000.0,
			HighPrice:  2100.0,
			LowPrice:   1950.0,
			ClosePrice: 2050.0,
			Volume:     1000000,
		},
	}
}

// WithID sets the stock price ID
func (b *StockPriceBuilder) WithID(id string) *StockPriceBuilder {
	b.stockPrice.ID = id
	return b
}

// WithCode sets the stock code
func (b *StockPriceBuilder) WithCode(code string) *StockPriceBuilder {
	b.stockPrice.Code = code
	return b
}

// WithDate sets the price date
func (b *StockPriceBuilder) WithDate(date time.Time) *StockPriceBuilder {
	b.stockPrice.Date = date
	return b
}

// WithPrices sets all price values at once
func (b *StockPriceBuilder) WithPrices(open, high, low, close float64) *StockPriceBuilder {
	b.stockPrice.OpenPrice = open
	b.stockPrice.HighPrice = high
	b.stockPrice.LowPrice = low
	b.stockPrice.ClosePrice = close
	return b
}

// WithVolume sets the volume
func (b *StockPriceBuilder) WithVolume(volume int64) *StockPriceBuilder {
	b.stockPrice.Volume = volume
	return b
}

// Build returns the built stock price
func (b *StockPriceBuilder) Build() *dao.StockPrice {
	return b.stockPrice
}

// Insert inserts the stock price into the database
func (b *StockPriceBuilder) Insert(ctx context.Context, db boil.ContextExecutor) error {
	return b.stockPrice.Insert(ctx, db, boil.Infer())
}

