package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// Repositories groups all repository interfaces.
type Repositories struct {
	Stock     StockRepository
	Portfolio PortfolioRepository
}

// Transaction represents a database transaction with repositories.
type Transaction struct {
	db   *sql.DB
	tx   *sql.Tx
	repo *Repositories
}

// TransactionManager manages database transactions.
type TransactionManager interface {
	// WithTransaction executes a function within a database transaction
	WithTransaction(ctx context.Context, fn func(*Repositories) error) error
	
	// GetRepositories returns repositories without transaction
	GetRepositories() *Repositories
}

// transactionManagerImpl implements TransactionManager.
type transactionManagerImpl struct {
	db   *sql.DB
	repo *Repositories
}

// NewTransactionManager creates a new transaction manager.
func NewTransactionManager(db *sql.DB) TransactionManager {
	return &transactionManagerImpl{
		db: db,
		repo: &Repositories{
			Stock:     NewStockRepository(db),
			Portfolio: NewPortfolioRepository(db),
		},
	}
}

// WithTransaction executes a function within a database transaction.
func (tm *transactionManagerImpl) WithTransaction(ctx context.Context, fn func(*Repositories) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create repositories with transaction
	repos := &Repositories{
		Stock:     NewStockRepository(tx),
		Portfolio: NewPortfolioRepository(tx),
	}

	// Execute the function
	if err := fn(repos); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			// Log rollback error, but return original error
			// In production, you would use a proper logger here
			return err
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// GetRepositories returns repositories without transaction.
func (tm *transactionManagerImpl) GetRepositories() *Repositories {
	return tm.repo
}

// ExecutorWrapper wraps boil.ContextExecutor to ensure proper type.
type ExecutorWrapper struct {
	boil.ContextExecutor
}

// NewExecutorWrapper creates a new executor wrapper.
func NewExecutorWrapper(executor boil.ContextExecutor) *ExecutorWrapper {
	return &ExecutorWrapper{ContextExecutor: executor}
}