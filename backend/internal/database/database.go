package database

import (
	"fmt"
	"time"

	"stock-automation/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	conn *gorm.DB
}

func NewDB() (*DB, error) {
	// データベース接続設定
	dsn := "root:password@tcp(localhost:3309)/stock_automation?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci&sql_mode=''"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 接続プールの設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// UTF-8文字セットを明示的に設定
	if err := db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return nil, fmt.Errorf("failed to set charset: %w", err)
	}

	if err := db.Exec("SET character_set_client = utf8mb4").Error; err != nil {
		return nil, fmt.Errorf("failed to set character_set_client: %w", err)
	}

	if err := db.Exec("SET character_set_connection = utf8mb4").Error; err != nil {
		return nil, fmt.Errorf("failed to set character_set_connection: %w", err)
	}

	if err := db.Exec("SET character_set_results = utf8mb4").Error; err != nil {
		return nil, fmt.Errorf("failed to set character_set_results: %w", err)
	}

	return &DB{conn: db}, nil
}

func (db *DB) AutoMigrate() error {
	return db.conn.AutoMigrate(
		&models.StockPrice{},
		&models.TechnicalIndicator{},
		&models.Portfolio{},
		&models.WatchList{},
	)
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
