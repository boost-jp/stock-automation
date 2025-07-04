# Docker環境構築ガイド

## 概要
Docker ComposeでMySQL 8データベースを構築し、Go アプリケーションと連携する環境設定

## 前提条件

### 必要ソフトウェア
- **Docker Desktop**: 最新版
- **Docker Compose**: v2.0以上（Docker Desktopに含まれる）
- **Go**: 1.19以上

### システム要件
- **メモリ**: 4GB以上（Docker用に2GB確保）
- **ストレージ**: 10GB以上の空き容量

## Docker構成

### プロジェクト構造
```
stock-automation/
├── docker/
│   ├── docker-compose.yml
│   ├── mysql/
│   │   ├── init/
│   │   │   ├── 01_create_database.sql
│   │   │   ├── 02_create_tables.sql
│   │   │   └── 03_insert_sample_data.sql
│   │   └── conf/
│   │       └── my.cnf
│   └── .env
├── cmd/
│   └── main.go
└── internal/
    └── database/
        └── mysql.go
```

## Docker Compose設定

### `docker/docker-compose.yml`
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: stock_mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
      TZ: Asia/Tokyo
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./mysql/init:/docker-entrypoint-initdb.d
      - ./mysql/conf/my.cnf:/etc/mysql/conf.d/my.cnf
    command: --default-authentication-plugin=mysql_native_password
    networks:
      - stock_network

  phpmyadmin:
    image: phpmyadmin/phpmyadmin:latest
    container_name: stock_phpmyadmin
    restart: unless-stopped
    environment:
      PMA_HOST: mysql
      PMA_USER: ${MYSQL_USER}
      PMA_PASSWORD: ${MYSQL_PASSWORD}
      UPLOAD_LIMIT: 100M
    ports:
      - "8080:80"
    depends_on:
      - mysql
    networks:
      - stock_network

  redis:
    image: redis:7-alpine
    container_name: stock_redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    networks:
      - stock_network

volumes:
  mysql_data:
    driver: local
  redis_data:
    driver: local

networks:
  stock_network:
    driver: bridge
```

### 環境変数設定 `docker/.env`
```env
# MySQL設定
MYSQL_ROOT_PASSWORD=root_password_123
MYSQL_DATABASE=stock_automation
MYSQL_USER=stock_user
MYSQL_PASSWORD=stock_password_456

# タイムゾーン
TZ=Asia/Tokyo
```

### MySQL設定 `docker/mysql/conf/my.cnf`
```ini
[mysqld]
# 基本設定
default-authentication-plugin=mysql_native_password
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
default-time-zone='+09:00'

# パフォーマンス設定
innodb_buffer_pool_size=1G
innodb_log_file_size=256M
innodb_flush_method=O_DIRECT
innodb_file_per_table=1

# ログ設定
slow_query_log=1
slow_query_log_file=/var/log/mysql/slow.log
long_query_time=2

# セキュリティ設定
sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO

[mysql]
default-character-set=utf8mb4

[client]
default-character-set=utf8mb4
```

## データベース初期化スクリプト

### `docker/mysql/init/01_create_database.sql`
```sql
-- データベース作成（既に存在する場合はスキップ）
CREATE DATABASE IF NOT EXISTS stock_automation 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

USE stock_automation;

-- ユーザー権限設定
GRANT ALL PRIVILEGES ON stock_automation.* TO 'stock_user'@'%';
FLUSH PRIVILEGES;
```

### `docker/mysql/init/02_create_tables.sql`
```sql
USE stock_automation;

-- 株価データテーブル
CREATE TABLE IF NOT EXISTS stock_prices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    volume BIGINT NOT NULL DEFAULT 0,
    high DECIMAL(10,2) NOT NULL,
    low DECIMAL(10,2) NOT NULL,
    open DECIMAL(10,2) NOT NULL,
    close DECIMAL(10,2) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_timestamp (timestamp),
    INDEX idx_code_timestamp (code, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- テクニカル指標テーブル
CREATE TABLE IF NOT EXISTS technical_indicators (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    ma5 DECIMAL(10,2),
    ma25 DECIMAL(10,2),
    ma75 DECIMAL(10,2),
    rsi DECIMAL(5,2),
    macd DECIMAL(10,4),
    signal DECIMAL(10,4),
    histogram DECIMAL(10,4),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_timestamp (timestamp),
    UNIQUE KEY unique_code_timestamp (code, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ポートフォリオテーブル
CREATE TABLE IF NOT EXISTS portfolios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    shares INT NOT NULL,
    purchase_price DECIMAL(10,2) NOT NULL,
    purchase_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 監視銘柄テーブル
CREATE TABLE IF NOT EXISTS watch_lists (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    target_buy_price DECIMAL(10,2),
    target_sell_price DECIMAL(10,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- アラート履歴テーブル
CREATE TABLE IF NOT EXISTS alert_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    alert_type ENUM('buy', 'sell', 'system') NOT NULL,
    message TEXT NOT NULL,
    price DECIMAL(10,2),
    confidence INT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_sent_at (sent_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### `docker/mysql/init/03_insert_sample_data.sql`
```sql
USE stock_automation;

-- サンプル監視銘柄データ
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price) VALUES
('7203', 'トヨタ自動車', 2500.00, 2800.00),
('6758', 'ソニーグループ', 12000.00, 15000.00),
('9984', 'ソフトバンクグループ', 5000.00, 6000.00),
('8306', '三菱UFJフィナンシャル・グループ', 800.00, 1000.00),
('9983', 'ファーストリテイリング', 80000.00, 90000.00)
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    target_buy_price = VALUES(target_buy_price),
    target_sell_price = VALUES(target_sell_price);

-- サンプルポートフォリオデータ
INSERT INTO portfolios (code, name, shares, purchase_price, purchase_date) VALUES
('7203', 'トヨタ自動車', 100, 2550.00, '2024-01-15'),
('6758', 'ソニーグループ', 50, 12500.00, '2024-02-10')
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    shares = VALUES(shares),
    purchase_price = VALUES(purchase_price),
    purchase_date = VALUES(purchase_date);
```

## Go MySQL接続設定

### 必要ライブラリのインストール
```bash
# MySQL ドライバー
go get github.com/go-sql-driver/mysql

# GORM ORM
go get gorm.io/gorm
go get gorm.io/driver/mysql

# Redis クライアント（キャッシュ用）
go get github.com/go-redis/redis/v8

# 設定管理
go get github.com/spf13/viper
```

### データベース接続設定 `internal/database/mysql.go`
```go
package database

import (
    "fmt"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "github.com/sirupsen/logrus"
)

type DB struct {
    conn *gorm.DB
}

type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
    Charset  string
    ParseTime bool
    Loc      string
}

func NewConnection(config Config) (*DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
        config.User,
        config.Password,
        config.Host,
        config.Port,
        config.Database,
        config.Charset,
        config.ParseTime,
        config.Loc,
    )
    
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    }
    
    conn, err := gorm.Open(mysql.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // 接続プール設定
    sqlDB, err := conn.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }
    
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    logrus.Info("Successfully connected to MySQL database")
    
    return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
    sqlDB, err := db.conn.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}

func (db *DB) GetConnection() *gorm.DB {
    return db.conn
}

// ヘルスチェック
func (db *DB) Ping() error {
    sqlDB, err := db.conn.DB()
    if err != nil {
        return err
    }
    return sqlDB.Ping()
}
```

### 設定ファイル更新 `configs/config.yaml`
```yaml
# データベース設定
database:
  host: "localhost"
  port: 3306
  user: "stock_user"
  password: "stock_password_456"
  database: "stock_automation"
  charset: "utf8mb4"
  parse_time: true
  loc: "Asia%2FTokyo"
  
# Redis設定
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
```

## 環境構築手順

### Step 1: Docker環境準備
```bash
# プロジェクトディレクトリ作成
mkdir stock-automation
cd stock-automation

# Docker設定ディレクトリ作成
mkdir -p docker/mysql/{init,conf}

# 上記のファイルを作成
# - docker-compose.yml
# - .env
# - MySQL設定ファイル
# - 初期化SQLスクリプト
```

### Step 2: Docker起動
```bash
# Docker Composeでサービス起動
cd docker
docker-compose up -d

# ログ確認
docker-compose logs -f mysql

# サービス状態確認
docker-compose ps
```

### Step 3: データベース接続確認
```bash
# MySQL接続テスト
docker exec -it stock_mysql mysql -u stock_user -p stock_automation

# または
mysql -h localhost -P 3306 -u stock_user -p stock_automation
```

### Step 4: Go アプリケーション接続テスト
```go
// test/db_connection_test.go
package main

import (
    "stock-automation/internal/database"
    "github.com/sirupsen/logrus"
)

func main() {
    config := database.Config{
        Host:      "localhost",
        Port:      3306,
        User:      "stock_user",
        Password:  "stock_password_456",
        Database:  "stock_automation",
        Charset:   "utf8mb4",
        ParseTime: true,
        Loc:       "Asia%2FTokyo",
    }
    
    db, err := database.NewConnection(config)
    if err != nil {
        logrus.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        logrus.Fatal("Database ping failed:", err)
    }
    
    logrus.Info("Database connection successful!")
}
```

## 管理ツール

### phpMyAdmin アクセス
- **URL**: http://localhost:8080
- **ユーザー**: stock_user
- **パスワード**: stock_password_456

### MySQL直接接続
```bash
# Docker経由
docker exec -it stock_mysql mysql -u stock_user -p

# ローカルから
mysql -h localhost -P 3306 -u stock_user -p stock_automation
```

### データベースバックアップ
```bash
# バックアップ作成
docker exec stock_mysql mysqldump -u root -p stock_automation > backup_$(date +%Y%m%d_%H%M%S).sql

# リストア
docker exec -i stock_mysql mysql -u root -p stock_automation < backup_20240101_120000.sql
```

## トラブルシューティング

### 一般的な問題

#### 1. ポート競合
```bash
# ポート使用状況確認
netstat -an | grep 3306
lsof -i :3306

# ポート変更（docker-compose.yml）
ports:
  - "3307:3306"  # ホストポートを変更
```

#### 2. 権限エラー
```bash
# MySQLコンテナ内で権限確認
docker exec -it stock_mysql mysql -u root -p
SHOW GRANTS FOR 'stock_user'@'%';

# 権限再設定
GRANT ALL PRIVILEGES ON stock_automation.* TO 'stock_user'@'%';
FLUSH PRIVILEGES;
```

#### 3. 文字化け問題
```bash
# 文字セット確認
SHOW VARIABLES LIKE 'character_set%';
SHOW VARIABLES LIKE 'collation%';

# 設定確認
SELECT @@character_set_database, @@collation_database;
```

#### 4. Docker容量不足
```bash
# 不要なイメージ・コンテナ削除
docker system prune -a

# ボリューム削除（注意：データが消える）
docker-compose down -v
```

### ログ確認
```bash
# MySQL ログ
docker-compose logs mysql

# エラーログ詳細
docker exec stock_mysql tail -f /var/log/mysql/error.log

# スロークエリログ
docker exec stock_mysql tail -f /var/log/mysql/slow.log
```

これでMySQL 8 + Docker Compose環境が構築できます。次は[データ収集システム](data-collection.md)をMySQL対応版に更新します。