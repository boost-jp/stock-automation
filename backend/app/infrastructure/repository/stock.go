package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
)

// StockRepository defines stock price related operations.
type StockRepository interface {
	// Stock price operations
	SaveStockPrice(ctx context.Context, price *models.StockPrice) error
	SaveStockPrices(ctx context.Context, prices []*models.StockPrice) error
	GetLatestPrice(ctx context.Context, stockCode string) (*models.StockPrice, error)
	GetPriceHistory(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error)
	CleanupOldData(ctx context.Context, days int) error

	// Technical indicator operations
	SaveTechnicalIndicator(ctx context.Context, indicator *models.TechnicalIndicator) error
	GetLatestTechnicalIndicator(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error)

	// Watch list operations
	GetActiveWatchList(ctx context.Context) ([]*models.WatchList, error)
	GetWatchListItem(ctx context.Context, id string) (*models.WatchList, error)
	AddToWatchList(ctx context.Context, item *models.WatchList) error
	UpdateWatchList(ctx context.Context, item *models.WatchList) error
	DeleteFromWatchList(ctx context.Context, id string) error
}

// stockRepositoryImpl implements StockRepository using SQLBoiler.
type stockRepositoryImpl struct {
	db boil.ContextExecutor
}

// NewStockRepository creates a new stock repository.
func NewStockRepository(db boil.ContextExecutor) StockRepository {
	return &stockRepositoryImpl{db: db}
}

// SaveStockPrice saves a single stock price record.
func (r *stockRepositoryImpl) SaveStockPrice(ctx context.Context, price *models.StockPrice) error {
	// Convert domain model to DAO model
	daoPrice := &dao.StockPrice{
		Code:       price.Code,
		Date:       price.Date,
		OpenPrice:  price.OpenPrice,
		HighPrice:  price.HighPrice,
		LowPrice:   price.LowPrice,
		ClosePrice: price.ClosePrice,
		Volume:     price.Volume,
	}

	return daoPrice.Insert(ctx, r.db, boil.Infer())
}

// SaveStockPrices saves multiple stock price records in batches.
func (r *stockRepositoryImpl) SaveStockPrices(ctx context.Context, prices []*models.StockPrice) error {
	if len(prices) == 0 {
		return nil
	}

	// Convert domain models to DAO models
	daoPrices := make(dao.StockPriceSlice, len(prices))
	for i, price := range prices {
		daoPrices[i] = &dao.StockPrice{
			Code:       price.Code,
			Date:       price.Date,
			OpenPrice:  price.OpenPrice,
			HighPrice:  price.HighPrice,
			LowPrice:   price.LowPrice,
			ClosePrice: price.ClosePrice,
			Volume:     price.Volume,
		}
	}

	return daoPrices.InsertAll(ctx, r.db, boil.Infer())
}

// GetLatestPrice retrieves the latest stock price for a given stock code.
func (r *stockRepositoryImpl) GetLatestPrice(ctx context.Context, stockCode string) (*models.StockPrice, error) {
	daoPrice, err := dao.StockPrices(
		qm.Where("code = ?", stockCode),
		qm.OrderBy("date desc"),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert DAO model to domain model
	return &models.StockPrice{
		Code:       daoPrice.Code,
		Date:       daoPrice.Date,
		OpenPrice:  daoPrice.OpenPrice,
		HighPrice:  daoPrice.HighPrice,
		LowPrice:   daoPrice.LowPrice,
		ClosePrice: daoPrice.ClosePrice,
		Volume:     daoPrice.Volume,
	}, nil
}

// GetPriceHistory retrieves stock price history for a given period.
func (r *stockRepositoryImpl) GetPriceHistory(ctx context.Context, stockCode string, days int) ([]*models.StockPrice, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	daoPrices, err := dao.StockPrices(
		qm.Where("code = ? AND date >= ?", stockCode, startTime),
		qm.OrderBy("date asc"),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	// Convert DAO models to domain models
	prices := make([]*models.StockPrice, len(daoPrices))
	for i, daoPrice := range daoPrices {
		prices[i] = &models.StockPrice{
			Code:       daoPrice.Code,
			Date:       daoPrice.Date,
			OpenPrice:  daoPrice.OpenPrice,
			HighPrice:  daoPrice.HighPrice,
			LowPrice:   daoPrice.LowPrice,
			ClosePrice: daoPrice.ClosePrice,
			Volume:     daoPrice.Volume,
		}
	}

	return prices, nil
}

// CleanupOldData removes old stock price data to manage database size.
func (r *stockRepositoryImpl) CleanupOldData(ctx context.Context, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	_, err := dao.StockPrices(
		qm.Where("date < ?", cutoffTime),
	).DeleteAll(ctx, r.db)

	return err
}

// SaveTechnicalIndicator saves a technical indicator record.
func (r *stockRepositoryImpl) SaveTechnicalIndicator(ctx context.Context, indicator *models.TechnicalIndicator) error {
	// Convert domain model to DAO model
	daoIndicator := &dao.TechnicalIndicator{
		Code:          indicator.Code,
		Date:          indicator.Date,
		Sma5:          indicator.Sma5,
		Sma25:         indicator.Sma25,
		Sma75:         indicator.Sma75,
		Rsi14:         indicator.Rsi14,
		Macd:          indicator.Macd,
		MacdSignal:    indicator.MacdSignal,
		MacdHistogram: indicator.MacdHistogram,
	}

	return daoIndicator.Insert(ctx, r.db, boil.Infer())
}

// GetLatestTechnicalIndicator retrieves the latest technical indicator for a given stock code.
func (r *stockRepositoryImpl) GetLatestTechnicalIndicator(ctx context.Context, stockCode string) (*models.TechnicalIndicator, error) {
	daoIndicator, err := dao.TechnicalIndicators(
		qm.Where("code = ?", stockCode),
		qm.OrderBy("date desc"),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert DAO model to domain model
	return &models.TechnicalIndicator{
		Code:          daoIndicator.Code,
		Date:          daoIndicator.Date,
		Sma5:          daoIndicator.Sma5,
		Sma25:         daoIndicator.Sma25,
		Sma75:         daoIndicator.Sma75,
		Rsi14:         daoIndicator.Rsi14,
		Macd:          daoIndicator.Macd,
		MacdSignal:    daoIndicator.MacdSignal,
		MacdHistogram: daoIndicator.MacdHistogram,
	}, nil
}

// GetActiveWatchList retrieves all active watch list items.
func (r *stockRepositoryImpl) GetActiveWatchList(ctx context.Context) ([]*models.WatchList, error) {
	daoWatchList, err := dao.WatchLists(
		qm.Where("is_active = ?", true),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	// Convert DAO models to domain models
	watchList := make([]*models.WatchList, len(daoWatchList))
	for i, daoItem := range daoWatchList {
		watchList[i] = &models.WatchList{
			ID:              daoItem.ID,
			Code:            daoItem.Code,
			Name:            daoItem.Name,
			TargetBuyPrice:  daoItem.TargetBuyPrice,
			TargetSellPrice: daoItem.TargetSellPrice,
			IsActive:        daoItem.IsActive,
			CreatedAt:       daoItem.CreatedAt,
			UpdatedAt:       daoItem.UpdatedAt,
		}
	}

	return watchList, nil
}

// GetWatchListItem retrieves a watch list item by ID.
func (r *stockRepositoryImpl) GetWatchListItem(ctx context.Context, id string) (*models.WatchList, error) {
	daoItem, err := dao.FindWatchList(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &models.WatchList{
		ID:              daoItem.ID,
		Code:            daoItem.Code,
		Name:            daoItem.Name,
		TargetBuyPrice:  daoItem.TargetBuyPrice,
		TargetSellPrice: daoItem.TargetSellPrice,
		IsActive:        daoItem.IsActive,
		CreatedAt:       daoItem.CreatedAt,
		UpdatedAt:       daoItem.UpdatedAt,
	}, nil
}

// AddToWatchList adds a new item to the watch list.
func (r *stockRepositoryImpl) AddToWatchList(ctx context.Context, item *models.WatchList) error {
	daoItem := &dao.WatchList{
		ID:              item.ID,
		Code:            item.Code,
		Name:            item.Name,
		TargetBuyPrice:  item.TargetBuyPrice,
		TargetSellPrice: item.TargetSellPrice,
		IsActive:        item.IsActive,
	}

	return daoItem.Insert(ctx, r.db, boil.Infer())
}

// UpdateWatchList updates an existing watch list item.
func (r *stockRepositoryImpl) UpdateWatchList(ctx context.Context, item *models.WatchList) error {
	daoItem := &dao.WatchList{
		ID:              item.ID,
		Code:            item.Code,
		Name:            item.Name,
		TargetBuyPrice:  item.TargetBuyPrice,
		TargetSellPrice: item.TargetSellPrice,
		IsActive:        item.IsActive,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}

	_, err := daoItem.Update(ctx, r.db, boil.Infer())
	return err
}

// DeleteFromWatchList removes an item from the watch list.
func (r *stockRepositoryImpl) DeleteFromWatchList(ctx context.Context, id string) error {
	daoItem := &dao.WatchList{ID: id}
	_, err := daoItem.Delete(ctx, r.db)
	return err
}
