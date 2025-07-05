# MySQL データベース設計

## 概要
株式投資自動化システムで使用するMySQL 8データベースのスキーマ設計と運用方法

## データベース構成

### 基本情報
- **データベース名**: stock_automation
- **文字セット**: utf8mb4
- **照合順序**: utf8mb4_unicode_ci
- **タイムゾーン**: Asia/Tokyo
- **エンジン**: InnoDB

## テーブル設計

### 1. 株価データテーブル (stock_prices)

#### 用途
リアルタイム・履歴株価データの保存

#### スキーマ
```sql
CREATE TABLE stock_prices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    name VARCHAR(100) NOT NULL COMMENT '銘柄名',
    price DECIMAL(10,2) NOT NULL COMMENT '現在価格',
    volume BIGINT NOT NULL DEFAULT 0 COMMENT '出来高',
    high DECIMAL(10,2) NOT NULL COMMENT '高値',
    low DECIMAL(10,2) NOT NULL COMMENT '安値',
    open DECIMAL(10,2) NOT NULL COMMENT '始値',
    close DECIMAL(10,2) NOT NULL COMMENT '終値',
    timestamp TIMESTAMP NOT NULL COMMENT 'データ取得時刻',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'レコード作成時刻',
    
    INDEX idx_code (code),
    INDEX idx_timestamp (timestamp),
    INDEX idx_code_timestamp (code, timestamp),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='株価データテーブル';
```

#### Go構造体
```go
type StockPrice struct {
    ID        uint      `gorm:"primaryKey;column:id"`
    Code      string    `gorm:"column:code;type:varchar(10);not null;index"`
    Name      string    `gorm:"column:name;type:varchar(100);not null"`
    Price     float64   `gorm:"column:price;type:decimal(10,2);not null"`
    Volume    int64     `gorm:"column:volume;type:bigint;not null;default:0"`
    High      float64   `gorm:"column:high;type:decimal(10,2);not null"`
    Low       float64   `gorm:"column:low;type:decimal(10,2);not null"`
    Open      float64   `gorm:"column:open;type:decimal(10,2);not null"`
    Close     float64   `gorm:"column:close;type:decimal(10,2);not null"`
    Timestamp time.Time `gorm:"column:timestamp;type:timestamp;not null;index"`
    CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (StockPrice) TableName() string {
    return "stock_prices"
}
```

### 2. テクニカル指標テーブル (technical_indicators)

#### 用途
移動平均線、RSI、MACD等のテクニカル指標保存

#### スキーマ
```sql
CREATE TABLE technical_indicators (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    ma5 DECIMAL(10,2) COMMENT '5日移動平均',
    ma25 DECIMAL(10,2) COMMENT '25日移動平均',
    ma75 DECIMAL(10,2) COMMENT '75日移動平均',
    rsi DECIMAL(5,2) COMMENT 'RSI指数',
    macd DECIMAL(10,4) COMMENT 'MACD値',
    signal DECIMAL(10,4) COMMENT 'シグナル値',
    histogram DECIMAL(10,4) COMMENT 'ヒストグラム値',
    bollinger_upper DECIMAL(10,2) COMMENT 'ボリンジャーバンド上限',
    bollinger_lower DECIMAL(10,2) COMMENT 'ボリンジャーバンド下限',
    timestamp TIMESTAMP NOT NULL COMMENT '計算基準時刻',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'レコード作成時刻',
    
    INDEX idx_code (code),
    INDEX idx_timestamp (timestamp),
    UNIQUE KEY unique_code_timestamp (code, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='テクニカル指標テーブル';
```

#### Go構造体
```go
type TechnicalIndicator struct {
    ID              uint      `gorm:"primaryKey"`
    Code            string    `gorm:"column:code;type:varchar(10);not null;index"`
    MA5             *float64  `gorm:"column:ma5;type:decimal(10,2)"`
    MA25            *float64  `gorm:"column:ma25;type:decimal(10,2)"`
    MA75            *float64  `gorm:"column:ma75;type:decimal(10,2)"`
    RSI             *float64  `gorm:"column:rsi;type:decimal(5,2)"`
    MACD            *float64  `gorm:"column:macd;type:decimal(10,4)"`
    Signal          *float64  `gorm:"column:signal;type:decimal(10,4)"`
    Histogram       *float64  `gorm:"column:histogram;type:decimal(10,4)"`
    BollingerUpper  *float64  `gorm:"column:bollinger_upper;type:decimal(10,2)"`
    BollingerLower  *float64  `gorm:"column:bollinger_lower;type:decimal(10,2)"`
    Timestamp       time.Time `gorm:"column:timestamp;type:timestamp;not null"`
    CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
}
```

### 3. ポートフォリオテーブル (portfolios)

#### 用途
保有銘柄の管理

#### スキーマ
```sql
CREATE TABLE portfolios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    name VARCHAR(100) NOT NULL COMMENT '銘柄名',
    shares INT NOT NULL COMMENT '保有株数',
    purchase_price DECIMAL(10,2) NOT NULL COMMENT '取得単価',
    purchase_date DATE NOT NULL COMMENT '取得日',
    notes TEXT COMMENT '備考',
    is_active BOOLEAN DEFAULT TRUE COMMENT '有効フラグ',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_code (code),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='ポートフォリオテーブル';
```

#### Go構造体
```go
type Portfolio struct {
    ID            uint      `gorm:"primaryKey"`
    Code          string    `gorm:"column:code;type:varchar(10);not null;index"`
    Name          string    `gorm:"column:name;type:varchar(100);not null"`
    Shares        int       `gorm:"column:shares;type:int;not null"`
    PurchasePrice float64   `gorm:"column:purchase_price;type:decimal(10,2);not null"`
    PurchaseDate  time.Time `gorm:"column:purchase_date;type:date;not null"`
    Notes         string    `gorm:"column:notes;type:text"`
    IsActive      bool      `gorm:"column:is_active;type:boolean;default:true"`
    CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
```

### 4. 監視銘柄テーブル (watch_lists)

#### 用途
価格監視対象銘柄の管理

#### スキーマ
```sql
CREATE TABLE watch_lists (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE COMMENT '銘柄コード',
    name VARCHAR(100) NOT NULL COMMENT '銘柄名',
    target_buy_price DECIMAL(10,2) COMMENT '目標買い価格',
    target_sell_price DECIMAL(10,2) COMMENT '目標売り価格',
    stop_loss_price DECIMAL(10,2) COMMENT '損切り価格',
    alert_enabled BOOLEAN DEFAULT TRUE COMMENT 'アラート有効',
    is_active BOOLEAN DEFAULT TRUE COMMENT '監視有効',
    priority INT DEFAULT 1 COMMENT '優先度(1:高, 2:中, 3:低)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_code (code),
    INDEX idx_is_active (is_active),
    INDEX idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='監視銘柄テーブル';
```

#### Go構造体
```go
type WatchList struct {
    ID              uint      `gorm:"primaryKey"`
    Code            string    `gorm:"column:code;type:varchar(10);not null;uniqueIndex"`
    Name            string    `gorm:"column:name;type:varchar(100);not null"`
    TargetBuyPrice  *float64  `gorm:"column:target_buy_price;type:decimal(10,2)"`
    TargetSellPrice *float64  `gorm:"column:target_sell_price;type:decimal(10,2)"`
    StopLossPrice   *float64  `gorm:"column:stop_loss_price;type:decimal(10,2)"`
    AlertEnabled    bool      `gorm:"column:alert_enabled;type:boolean;default:true"`
    IsActive        bool      `gorm:"column:is_active;type:boolean;default:true"`
    Priority        int       `gorm:"column:priority;type:int;default:1"`
    CreatedAt       time.Time `gorm:"autoCreateTime"`
    UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}
```

### 5. アラート履歴テーブル (alert_history)

#### 用途
送信されたアラートの履歴管理

#### スキーマ
```sql
CREATE TABLE alert_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    alert_type ENUM('buy', 'sell', 'system', 'error') NOT NULL COMMENT 'アラート種類',
    message TEXT NOT NULL COMMENT 'アラートメッセージ',
    price DECIMAL(10,2) COMMENT '価格',
    confidence INT COMMENT '信頼度(%)',
    channel VARCHAR(50) DEFAULT 'slack' COMMENT '送信チャネル',
    status ENUM('sent', 'failed', 'pending') DEFAULT 'sent' COMMENT '送信状態',
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '送信日時',
    
    INDEX idx_code (code),
    INDEX idx_alert_type (alert_type),
    INDEX idx_sent_at (sent_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='アラート履歴テーブル';
```

#### Go構造体
```go
type AlertHistory struct {
    ID         uint      `gorm:"primaryKey"`
    Code       string    `gorm:"column:code;type:varchar(10);not null;index"`
    AlertType  string    `gorm:"column:alert_type;type:enum('buy','sell','system','error');not null"`
    Message    string    `gorm:"column:message;type:text;not null"`
    Price      *float64  `gorm:"column:price;type:decimal(10,2)"`
    Confidence *int      `gorm:"column:confidence;type:int"`
    Channel    string    `gorm:"column:channel;type:varchar(50);default:'slack'"`
    Status     string    `gorm:"column:status;type:enum('sent','failed','pending');default:'sent'"`
    SentAt     time.Time `gorm:"column:sent_at;type:timestamp;default:CURRENT_TIMESTAMP"`
}
```

### 6. システム設定テーブル (system_settings)

#### 用途
アプリケーション設定の永続化

#### スキーマ
```sql
CREATE TABLE system_settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    setting_key VARCHAR(100) NOT NULL UNIQUE COMMENT '設定キー',
    setting_value TEXT COMMENT '設定値',
    description TEXT COMMENT '説明',
    data_type ENUM('string', 'int', 'float', 'boolean', 'json') DEFAULT 'string',
    is_active BOOLEAN DEFAULT TRUE COMMENT '有効フラグ',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_setting_key (setting_key),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='システム設定テーブル';
```

## データベース操作

### 基本操作クラス

#### `app/database/operations.go`
```go
package database

import (
    "time"
    "gorm.io/gorm"
    "stock-automation/app/models"
)

type StockOperations struct {
    db *gorm.DB
}

func NewStockOperations(db *gorm.DB) *StockOperations {
    return &StockOperations{db: db}
}

// 株価データ操作
func (s *StockOperations) SaveStockPrice(price *models.StockPrice) error {
    return s.db.Create(price).Error
}

func (s *StockOperations) SaveStockPrices(prices []models.StockPrice) error {
    return s.db.CreateInBatches(prices, 100).Error
}

func (s *StockOperations) GetLatestPrice(stockCode string) (*models.StockPrice, error) {
    var price models.StockPrice
    err := s.db.Where("code = ?", stockCode).
        Order("timestamp DESC").
        First(&price).Error
    
    if err != nil {
        return nil, err
    }
    return &price, nil
}

func (s *StockOperations) GetPriceHistory(stockCode string, days int) ([]models.StockPrice, error) {
    var prices []models.StockPrice
    startTime := time.Now().AddDate(0, 0, -days)
    
    err := s.db.Where("code = ? AND timestamp >= ?", stockCode, startTime).
        Order("timestamp ASC").
        Find(&prices).Error
    
    return prices, err
}

// ポートフォリオ操作
func (s *StockOperations) GetActivePortfolio() ([]models.Portfolio, error) {
    var portfolio []models.Portfolio
    err := s.db.Where("is_active = ?", true).Find(&portfolio).Error
    return portfolio, err
}

func (s *StockOperations) AddToPortfolio(item *models.Portfolio) error {
    return s.db.Create(item).Error
}

func (s *StockOperations) UpdatePortfolioItem(id uint, updates map[string]interface{}) error {
    return s.db.Model(&models.Portfolio{}).Where("id = ?", id).Updates(updates).Error
}

// 監視銘柄操作
func (s *StockOperations) GetActiveWatchList() ([]models.WatchList, error) {
    var watchList []models.WatchList
    err := s.db.Where("is_active = ?", true).
        Order("priority ASC, code ASC").
        Find(&watchList).Error
    return watchList, err
}

func (s *StockOperations) AddToWatchList(item *models.WatchList) error {
    return s.db.Create(item).Error
}

// アラート履歴操作
func (s *StockOperations) SaveAlertHistory(alert *models.AlertHistory) error {
    return s.db.Create(alert).Error
}

func (s *StockOperations) GetRecentAlerts(hours int) ([]models.AlertHistory, error) {
    var alerts []models.AlertHistory
    since := time.Now().Add(-time.Duration(hours) * time.Hour)
    
    err := s.db.Where("sent_at >= ?", since).
        Order("sent_at DESC").
        Find(&alerts).Error
    
    return alerts, err
}

// データクリーンアップ
func (s *StockOperations) CleanupOldData(days int) error {
    cutoffTime := time.Now().AddDate(0, 0, -days)
    
    // 古い株価データ削除
    if err := s.db.Where("timestamp < ?", cutoffTime).Delete(&models.StockPrice{}).Error; err != nil {
        return err
    }
    
    // 古いテクニカル指標削除
    if err := s.db.Where("timestamp < ?", cutoffTime).Delete(&models.TechnicalIndicator{}).Error; err != nil {
        return err
    }
    
    // 古いアラート履歴削除（1年以上）
    alertCutoff := time.Now().AddDate(-1, 0, 0)
    if err := s.db.Where("sent_at < ?", alertCutoff).Delete(&models.AlertHistory{}).Error; err != nil {
        return err
    }
    
    return nil
}
```

## インデックス戦略

### パフォーマンス最適化
```sql
-- 複合インデックス（よく使用される検索パターン）
CREATE INDEX idx_stock_code_timestamp ON stock_prices(code, timestamp DESC);
CREATE INDEX idx_watch_active_priority ON watch_lists(is_active, priority);
CREATE INDEX idx_alert_type_date ON alert_history(alert_type, sent_at DESC);

-- カバリングインデックス（SELECT専用）
CREATE INDEX idx_price_summary ON stock_prices(code, timestamp, price, volume);
```

### クエリ最適化例
```sql
-- 最新価格取得（インデックス活用）
SELECT price, volume, timestamp 
FROM stock_prices 
WHERE code = '7203' 
ORDER BY timestamp DESC 
LIMIT 1;

-- 期間別データ取得
SELECT * FROM stock_prices 
WHERE code = '7203' 
  AND timestamp >= DATE_SUB(NOW(), INTERVAL 30 DAY)
ORDER BY timestamp ASC;
```

## バックアップ・復旧

### 自動バックアップスクリプト
```bash
#!/bin/bash
# scripts/backup_db.sh

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="stock_automation_backup_${DATE}.sql"

mkdir -p $BACKUP_DIR

# データベースバックアップ
docker exec stock_mysql mysqldump \
  -u stock_user \
  -pstock_password_456 \
  --single-transaction \
  --routines \
  --triggers \
  stock_automation > "${BACKUP_DIR}/${BACKUP_FILE}"

# 圧縮
gzip "${BACKUP_DIR}/${BACKUP_FILE}"

# 古いバックアップ削除（30日以上）
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
```

### 復旧手順
```bash
# バックアップから復旧
docker exec -i stock_mysql mysql -u root -p stock_automation < backup_file.sql

# 特定テーブルのみ復旧
docker exec -i stock_mysql mysql -u root -p stock_automation \
  -e "DROP TABLE IF EXISTS stock_prices;"
docker exec -i stock_mysql mysql -u root -p stock_automation < stock_prices_backup.sql
```

## 監視・メンテナンス

### パフォーマンス監視クエリ
```sql
-- スロークエリ確認
SELECT query_time, lock_time, rows_sent, rows_examined, sql_text 
FROM mysql.slow_log 
ORDER BY query_time DESC 
LIMIT 10;

-- インデックス使用状況
SELECT OBJECT_SCHEMA, OBJECT_NAME, INDEX_NAME, COUNT_FETCH, COUNT_INSERT, COUNT_UPDATE, COUNT_DELETE
FROM performance_schema.table_io_waits_summary_by_index_usage 
WHERE OBJECT_SCHEMA = 'stock_automation'
ORDER BY COUNT_FETCH DESC;

-- テーブルサイズ確認
SELECT 
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)'
FROM information_schema.tables 
WHERE table_schema = 'stock_automation'
ORDER BY (data_length + index_length) DESC;
```

この設計により、高性能で拡張性のあるMySQL データベースシステムが構築できます。