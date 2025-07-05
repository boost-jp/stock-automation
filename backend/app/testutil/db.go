package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
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
		Port:         getEnvOrDefaultInt("TEST_DB_PORT", 3306),
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
		dao.TableNames.Portfolios,
		dao.TableNames.StockPrices,
		dao.TableNames.TechnicalIndicators,
		dao.TableNames.WatchLists,
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
	// Read the schema from the main schema.sql file
	schemaPath := "schema.sql"
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		// If schema.sql is not found, use the embedded schema
		return createEmbeddedTestSchema(db)
	}

	// Execute schema creation
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("failed to create schema from file: %w", err)
	}

	return nil
}

// createEmbeddedTestSchema creates the database schema using embedded SQL
func createEmbeddedTestSchema(db *sql.DB) error {
	// Create tables one by one to avoid SQL syntax errors
	tables := []string{
		`CREATE TABLE IF NOT EXISTS watch_lists (
			id VARCHAR(26) PRIMARY KEY,
			code VARCHAR(10) NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			target_buy_price DECIMAL(10,2),
			target_sell_price DECIMAL(10,2),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS stock_prices (
			id VARCHAR(26) PRIMARY KEY,
			code VARCHAR(10) NOT NULL,
			date DATE NOT NULL,
			open_price DECIMAL(10,2) NOT NULL,
			high_price DECIMAL(10,2) NOT NULL,
			low_price DECIMAL(10,2) NOT NULL,
			close_price DECIMAL(10,2) NOT NULL,
			volume BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_code_date (code, date),
			INDEX idx_code (code),
			INDEX idx_date (date)
		)`,
		`CREATE TABLE IF NOT EXISTS technical_indicators (
			id VARCHAR(26) PRIMARY KEY,
			code VARCHAR(10) NOT NULL,
			date DATE NOT NULL,
			rsi_14 DECIMAL(5,2),
			macd DECIMAL(10,4),
			macd_signal DECIMAL(10,4),
			macd_histogram DECIMAL(10,4),
			sma_5 DECIMAL(10,2),
			sma_25 DECIMAL(10,2),
			sma_75 DECIMAL(10,2),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_code_date (code, date),
			INDEX idx_code (code),
			INDEX idx_date (date)
		)`,
		`CREATE TABLE IF NOT EXISTS portfolios (
			id VARCHAR(26) PRIMARY KEY,
			code VARCHAR(10) NOT NULL,
			name VARCHAR(100) NOT NULL,
			shares INT NOT NULL,
			purchase_price DECIMAL(10,2) NOT NULL,
			purchase_date DATE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_code (code)
		)`,
	}

	// Execute each table creation separately
	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
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

// getEnvOrDefaultInt returns environment variable value as int or default
func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper functions for test data creation are now provided by the fixture package
// See app/testutil/fixture for test data builders
