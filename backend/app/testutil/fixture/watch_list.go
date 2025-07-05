package fixture

import (
	"context"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
	"github.com/oklog/ulid/v2"
)

// WatchListBuilder helps build test watch list data
type WatchListBuilder struct {
	watchList *dao.WatchList
}

// NewWatchList creates a new watch list builder with default values
func NewWatchList() *WatchListBuilder {
	return &WatchListBuilder{
		watchList: &dao.WatchList{
			ID:              ulid.MustNew(ulid.Now(), nil).String(),
			Code:            "7203",
			Name:            "トヨタ自動車",
			IsActive:        true,
			TargetBuyPrice:  null.Float64{},
			TargetSellPrice: null.Float64{},
		},
	}
}

// WithID sets the watch list ID
func (b *WatchListBuilder) WithID(id string) *WatchListBuilder {
	b.watchList.ID = id
	return b
}

// WithCode sets the stock code
func (b *WatchListBuilder) WithCode(code string) *WatchListBuilder {
	b.watchList.Code = code
	return b
}

// WithName sets the stock name
func (b *WatchListBuilder) WithName(name string) *WatchListBuilder {
	b.watchList.Name = name
	return b
}

// WithIsActive sets the active status
func (b *WatchListBuilder) WithIsActive(active bool) *WatchListBuilder {
	b.watchList.IsActive = active
	return b
}

// WithTargetBuyPrice sets the target buy price
func (b *WatchListBuilder) WithTargetBuyPrice(price float64) *WatchListBuilder {
	b.watchList.TargetBuyPrice = null.Float64From(price)
	return b
}

// WithTargetSellPrice sets the target sell price
func (b *WatchListBuilder) WithTargetSellPrice(price float64) *WatchListBuilder {
	b.watchList.TargetSellPrice = null.Float64From(price)
	return b
}

// Build returns the built watch list
func (b *WatchListBuilder) Build() *dao.WatchList {
	return b.watchList
}

// Insert inserts the watch list into the database
func (b *WatchListBuilder) Insert(ctx context.Context, db boil.ContextExecutor) error {
	return b.watchList.Insert(ctx, db, boil.Infer())
}