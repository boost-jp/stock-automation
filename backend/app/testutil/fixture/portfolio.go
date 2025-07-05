package fixture

import (
	"context"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
	"github.com/oklog/ulid/v2"
)

// PortfolioBuilder helps build test portfolio data
type PortfolioBuilder struct {
	portfolio *dao.Portfolio
}

// NewPortfolio creates a new portfolio builder with default values
func NewPortfolio() *PortfolioBuilder {
	return &PortfolioBuilder{
		portfolio: &dao.Portfolio{
			ID:            ulid.MustNew(ulid.Now(), nil).String(),
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: client.FloatToDecimal(2000.0),
			PurchaseDate:  time.Now(),
		},
	}
}

// WithID sets the portfolio ID
func (b *PortfolioBuilder) WithID(id string) *PortfolioBuilder {
	b.portfolio.ID = id
	return b
}

// WithCode sets the stock code
func (b *PortfolioBuilder) WithCode(code string) *PortfolioBuilder {
	b.portfolio.Code = code
	return b
}

// WithName sets the stock name
func (b *PortfolioBuilder) WithName(name string) *PortfolioBuilder {
	b.portfolio.Name = name
	return b
}

// WithShares sets the number of shares
func (b *PortfolioBuilder) WithShares(shares int) *PortfolioBuilder {
	b.portfolio.Shares = shares
	return b
}

// WithPurchasePrice sets the purchase price
func (b *PortfolioBuilder) WithPurchasePrice(price float64) *PortfolioBuilder {
	b.portfolio.PurchasePrice = client.FloatToDecimal(price)
	return b
}

// WithPurchaseDate sets the purchase date
func (b *PortfolioBuilder) WithPurchaseDate(date time.Time) *PortfolioBuilder {
	b.portfolio.PurchaseDate = date
	return b
}

// Build returns the built portfolio
func (b *PortfolioBuilder) Build() *dao.Portfolio {
	return b.portfolio
}

// Insert inserts the portfolio into the database
func (b *PortfolioBuilder) Insert(ctx context.Context, db boil.ContextExecutor) error {
	return b.portfolio.Insert(ctx, db, boil.Infer())
}
