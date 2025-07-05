package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
)

// PortfolioRepository defines portfolio related operations.
type PortfolioRepository interface {
	// Portfolio CRUD operations
	Create(ctx context.Context, portfolio *models.Portfolio) error
	GetByID(ctx context.Context, id string) (*models.Portfolio, error)
	GetByCode(ctx context.Context, code string) (*models.Portfolio, error)
	GetAll(ctx context.Context) ([]*models.Portfolio, error)
	Update(ctx context.Context, portfolio *models.Portfolio) error
	Delete(ctx context.Context, id string) error

	// Aggregate operations
	GetTotalValue(ctx context.Context, currentPrices map[string]float64) (float64, error)
	GetHoldingsByCode(ctx context.Context, codes []string) ([]*models.Portfolio, error)
}

// portfolioRepositoryImpl implements PortfolioRepository using SQLBoiler.
type portfolioRepositoryImpl struct {
	db boil.ContextExecutor
}

// NewPortfolioRepository creates a new portfolio repository.
func NewPortfolioRepository(db boil.ContextExecutor) PortfolioRepository {
	return &portfolioRepositoryImpl{db: db}
}

// Create creates a new portfolio record.
func (r *portfolioRepositoryImpl) Create(ctx context.Context, portfolio *models.Portfolio) error {
	// Convert domain model to DAO model
	daoPortfolio := &dao.Portfolio{
		ID:            portfolio.ID,
		Code:          portfolio.Code,
		Name:          portfolio.Name,
		Shares:        portfolio.Shares,
		PurchasePrice: portfolio.PurchasePrice,
		PurchaseDate:  portfolio.PurchaseDate,
	}

	err := daoPortfolio.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return err
	}

	// Update the domain model with generated values
	portfolio.CreatedAt = daoPortfolio.CreatedAt
	portfolio.UpdatedAt = daoPortfolio.UpdatedAt

	return nil
}

// GetByID retrieves a portfolio by its ID.
func (r *portfolioRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Portfolio, error) {
	daoPortfolio, err := dao.FindPortfolio(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.convertToModel(daoPortfolio), nil
}

// GetByCode retrieves a portfolio by stock code.
func (r *portfolioRepositoryImpl) GetByCode(ctx context.Context, code string) (*models.Portfolio, error) {
	daoPortfolio, err := dao.Portfolios(
		qm.Where("code = ?", code),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.convertToModel(daoPortfolio), nil
}

// GetAll retrieves all portfolio records.
func (r *portfolioRepositoryImpl) GetAll(ctx context.Context) ([]*models.Portfolio, error) {
	daoPortfolios, err := dao.Portfolios().All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	portfolios := make([]*models.Portfolio, len(daoPortfolios))
	for i, daoPortfolio := range daoPortfolios {
		portfolios[i] = r.convertToModel(daoPortfolio)
	}

	return portfolios, nil
}

// Update updates an existing portfolio record.
func (r *portfolioRepositoryImpl) Update(ctx context.Context, portfolio *models.Portfolio) error {
	// Convert domain model to DAO model
	daoPortfolio := &dao.Portfolio{
		ID:            portfolio.ID,
		Code:          portfolio.Code,
		Name:          portfolio.Name,
		Shares:        portfolio.Shares,
		PurchasePrice: portfolio.PurchasePrice,
		PurchaseDate:  portfolio.PurchaseDate,
		CreatedAt:     portfolio.CreatedAt,
		UpdatedAt:     portfolio.UpdatedAt,
	}

	_, err := daoPortfolio.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return err
	}

	// Update the domain model with new values
	portfolio.UpdatedAt = daoPortfolio.UpdatedAt

	return nil
}

// Delete removes a portfolio record by ID.
func (r *portfolioRepositoryImpl) Delete(ctx context.Context, id string) error {
	daoPortfolio := &dao.Portfolio{ID: id}
	_, err := daoPortfolio.Delete(ctx, r.db)
	return err
}

// GetTotalValue calculates the total value of all portfolios with current prices.
func (r *portfolioRepositoryImpl) GetTotalValue(ctx context.Context, currentPrices map[string]float64) (float64, error) {
	portfolios, err := r.GetAll(ctx)
	if err != nil {
		return 0, err
	}

	totalValue := 0.0
	for _, portfolio := range portfolios {
		if currentPrice, exists := currentPrices[portfolio.Code]; exists {
			totalValue += portfolio.CalculateCurrentValue(currentPrice)
		}
	}

	return totalValue, nil
}

// GetHoldingsByCode retrieves portfolios for specific stock codes.
func (r *portfolioRepositoryImpl) GetHoldingsByCode(ctx context.Context, codes []string) ([]*models.Portfolio, error) {
	if len(codes) == 0 {
		return []*models.Portfolio{}, nil
	}

	// Convert codes to interface{} slice for SQLBoiler
	args := make([]interface{}, len(codes))
	for i, code := range codes {
		args[i] = code
	}

	daoPortfolios, err := dao.Portfolios(
		qm.WhereIn("code IN ?", args...),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	portfolios := make([]*models.Portfolio, len(daoPortfolios))
	for i, daoPortfolio := range daoPortfolios {
		portfolios[i] = r.convertToModel(daoPortfolio)
	}

	return portfolios, nil
}

// convertToModel converts DAO portfolio to domain model.
func (r *portfolioRepositoryImpl) convertToModel(daoPortfolio *dao.Portfolio) *models.Portfolio {
	return &models.Portfolio{
		ID:            daoPortfolio.ID,
		Code:          daoPortfolio.Code,
		Name:          daoPortfolio.Name,
		Shares:        daoPortfolio.Shares,
		PurchasePrice: daoPortfolio.PurchasePrice,
		PurchaseDate:  daoPortfolio.PurchaseDate,
		CreatedAt:     daoPortfolio.CreatedAt,
		UpdatedAt:     daoPortfolio.UpdatedAt,
	}
}
