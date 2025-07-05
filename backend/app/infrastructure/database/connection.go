package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	_ "github.com/go-sql-driver/mysql"
)

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// ConnectionManager manages database connections.
type ConnectionManager interface {
	GetDB() *sql.DB
	GetExecutor() boil.ContextExecutor
	Close() error
	Ping() error
	GetStats() sql.DBStats
}

// connectionManagerImpl implements ConnectionManager.
type connectionManagerImpl struct {
	db *sql.DB
}

// NewConnectionManager creates a new database connection manager.
func NewConnectionManager(config DatabaseConfig) (ConnectionManager, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DatabaseName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.MaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &connectionManagerImpl{db: db}, nil
}

// GetDB returns the underlying sql.DB instance.
func (c *connectionManagerImpl) GetDB() *sql.DB {
	return c.db
}

// GetExecutor returns a boil.ContextExecutor for SQLBoiler operations.
func (c *connectionManagerImpl) GetExecutor() boil.ContextExecutor {
	return c.db
}

// Close closes the database connection.
func (c *connectionManagerImpl) Close() error {
	return c.db.Close()
}

// Ping checks if the database connection is alive.
func (c *connectionManagerImpl) Ping() error {
	return c.db.Ping()
}

// GetStats returns database connection statistics.
func (c *connectionManagerImpl) GetStats() sql.DBStats {
	return c.db.Stats()
}

// DefaultDatabaseConfig returns default database configuration.
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "",
		DatabaseName: "stock_automation",
		MaxOpenConns: 25,
		MaxIdleConns: 10,
		MaxLifetime:  5 * time.Minute,
	}
}
