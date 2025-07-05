package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	_ "github.com/go-sql-driver/mysql"
)

// TestDB represents a test database connection
type TestDB struct {
	DB             *sql.DB
	ConnectionMgr  database.ConnectionManager
	cleanupFuncs   []func() error
	testSchemaName string
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Use test-specific database configuration
	config := database.DatabaseConfig{
		Host:         getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:         3306,
		User:         getEnvOrDefault("TEST_DB_USER", "root"),
		Password:     getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		DatabaseName: fmt.Sprintf("test_stock_automation_%d", time.Now().Unix()),
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}

	// Create connection to MySQL without specifying database
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.User, config.Password, config.Host, config.Port)

	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer rootDB.Close()

	// Create test database
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", config.DatabaseName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	connMgr, err := database.NewConnectionManager(config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db := connMgr.GetDB()

	// Create schema
	if err := createTestSchema(db); err != nil {
		connMgr.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	testDB := &TestDB{
		DB:             db,
		ConnectionMgr:  connMgr,
		testSchemaName: config.DatabaseName,
		cleanupFuncs:   []func() error{},
	}

	// Add cleanup function to drop database
	testDB.cleanupFuncs = append(testDB.cleanupFuncs, func() error {
		_, err := rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", config.DatabaseName))
		return err
	})

	return testDB
}

// GetBoilDB returns a boil.ContextExecutor for SQLBoiler operations
func (tdb *TestDB) GetBoilDB() boil.ContextExecutor {
	return tdb.DB
}

// GetDB returns the underlying *sql.DB
func (tdb *TestDB) GetDB() *sql.DB {
	return tdb.DB
}

// Cleanup performs cleanup operations
func (tdb *TestDB) Cleanup() error {
	// Close connection first
	if err := tdb.ConnectionMgr.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	// Run cleanup functions in reverse order
	for i := len(tdb.cleanupFuncs) - 1; i >= 0; i-- {
		if err := tdb.cleanupFuncs[i](); err != nil {
			return fmt.Errorf("cleanup function %d failed: %w", i, err)
		}
	}

	return nil
}

// TruncateAll truncates all tables in the test database
func (tdb *TestDB) TruncateAll() error {
	tables := []string{
		"portfolios",
		"stock_prices",
		"technical_indicators",
		"watch_lists",
	}

	// Disable foreign key checks
	if _, err := tdb.DB.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		if _, err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", table)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := tdb.DB.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %w", err)
	}

	return nil
}

// ExecSQL executes raw SQL statements (useful for test data setup)
func (tdb *TestDB) ExecSQL(query string, args ...interface{}) error {
	_, err := tdb.DB.Exec(query, args...)
	return err
}

// WithTransaction runs a function within a database transaction
func (tdb *TestDB) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := tdb.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// createTestSchema creates the database schema for testing
func createTestSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS watch_lists (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stock_prices (
    code VARCHAR(10) NOT NULL,
    date DATE NOT NULL,
    open_price DECIMAL(10,2) NOT NULL,
    high_price DECIMAL(10,2) NOT NULL,
    low_price DECIMAL(10,2) NOT NULL,
    close_price DECIMAL(10,2) NOT NULL,
    volume BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (code, date),
    INDEX idx_date (date),
    INDEX idx_code_date (code, date)
);

CREATE TABLE IF NOT EXISTS technical_indicators (
    code VARCHAR(10) NOT NULL,
    date DATE NOT NULL,
    indicator_type VARCHAR(50) NOT NULL,
    value DECIMAL(20,6) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (code, date, indicator_type),
    INDEX idx_code_type (code, indicator_type),
    INDEX idx_date (date)
);

CREATE TABLE IF NOT EXISTS portfolios (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    shares INT NOT NULL,
    purchase_price DECIMAL(10,2) NOT NULL,
    purchase_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
);
`

	// Execute schema creation
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions for test data creation

// InsertTestPortfolio inserts a test portfolio record
func (tdb *TestDB) InsertTestPortfolio(ctx context.Context, id, code, name string, shares int, purchasePrice float64, purchaseDate time.Time) error {
	query := `
		INSERT INTO portfolios (id, code, name, shares, purchase_price, purchase_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	return tdb.ExecSQL(query, id, code, name, shares, purchasePrice, purchaseDate)
}

// InsertTestStockPrice inserts a test stock price record
func (tdb *TestDB) InsertTestStockPrice(ctx context.Context, code string, date time.Time, open, high, low, close float64, volume int64) error {
	query := `
		INSERT INTO stock_prices (code, date, open_price, high_price, low_price, close_price, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	return tdb.ExecSQL(query, code, date, open, high, low, close, volume)
}

// InsertTestWatchList inserts a test watch list record
func (tdb *TestDB) InsertTestWatchList(ctx context.Context, id, code, name string, isActive bool) error {
	query := `
		INSERT INTO watch_lists (id, code, name, is_active)
		VALUES (?, ?, ?, ?)
	`
	return tdb.ExecSQL(query, id, code, name, isActive)
}
