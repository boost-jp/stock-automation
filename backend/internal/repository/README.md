# Repository Layer

## Overview
This package implements the Repository pattern for data access layer using SQLBoiler ORM. It provides a clean abstraction over database operations and supports transactions.

## Architecture

```
Domain Layer (models)
        ↓
Repository Interface
        ↓
Repository Implementation (SQLBoiler)
        ↓
Database
```

## Components

### StockRepository
Handles all stock-related data operations:
- Stock price CRUD operations
- Technical indicator operations
- Watch list management
- Historical data queries

### PortfolioRepository
Manages portfolio data:
- Portfolio CRUD operations
- Aggregate calculations
- Holdings management

### TransactionManager
Provides transaction support:
- Atomic operations across multiple repositories
- Rollback on errors
- Connection management

## Usage

### Basic Operations

```go
// Create transaction manager
db, _ := sql.Open("mysql", connectionString)
tm := repository.NewTransactionManager(db)

// Get repositories (without transaction)
repos := tm.GetRepositories()

// Save stock price
ctx := context.Background()
stockPrice := &models.StockPrice{...}
err := repos.Stock.SaveStockPrice(ctx, stockPrice)

// Get portfolio
portfolio, err := repos.Portfolio.GetByCode(ctx, "1234")
```

### With Transactions

```go
// Execute multiple operations in a transaction
err := tm.WithTransaction(ctx, func(repos *repository.Repositories) error {
    // Save stock price
    if err := repos.Stock.SaveStockPrice(ctx, stockPrice); err != nil {
        return err
    }
    
    // Update portfolio
    if err := repos.Portfolio.Update(ctx, portfolio); err != nil {
        return err
    }
    
    return nil // Commit transaction
})
```

## Migration from Legacy Code

### Before (app/database/stock_operations.go)
```go
// Old way
func (db *DB) SaveStockPrice(price *models.StockPrice) error {
    return db.conn.Create(price).Error
}
```

### After (Repository Pattern)
```go
// New way
func (r *stockRepositoryImpl) SaveStockPrice(ctx context.Context, price *models.StockPrice) error {
    daoPrice := &dao.StockPrice{...}
    return daoPrice.Insert(ctx, r.db, boil.Infer())
}
```

### Migration Steps

1. **Replace direct database calls**: 
   - Old: `db.SaveStockPrice(price)`
   - New: `repos.Stock.SaveStockPrice(ctx, price)`

2. **Add context to all operations**:
   - All repository methods require `context.Context`

3. **Use transaction manager for multi-step operations**:
   - Wrap related operations in `WithTransaction`

4. **Update error handling**:
   - Repository methods return domain-appropriate errors

## Testing

The repository layer includes comprehensive tests:
- Interface compliance tests
- Mock-based unit tests
- Transaction behavior tests

To run tests:
```bash
go test ./internal/repository/...
```

## Benefits

1. **Clean Architecture**: Clear separation between domain and data access
2. **Testability**: Easy to mock and test business logic
3. **Transaction Support**: Built-in transaction management
4. **Type Safety**: Compile-time verification of data operations
5. **Consistency**: Standardized patterns across all data access