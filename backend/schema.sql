-- Stock Automation Database Schema

-- 株価データテーブル
CREATE TABLE stock_prices (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    `date` DATE NOT NULL COMMENT '取引日',
    open_price DECIMAL(10,2) NOT NULL COMMENT '始値',
    high_price DECIMAL(10,2) NOT NULL COMMENT '高値',
    low_price DECIMAL(10,2) NOT NULL COMMENT '安値',
    close_price DECIMAL(10,2) NOT NULL COMMENT '終値',
    volume BIGINT NOT NULL COMMENT '出来高',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    UNIQUE KEY unique_code_date (code, `date`),
    INDEX idx_code (code),
    INDEX idx_date (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='株価データ';

-- テクニカル指標テーブル
CREATE TABLE technical_indicators (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    `date` DATE NOT NULL COMMENT '計算日',
    rsi_14 DECIMAL(5,2) COMMENT 'RSI(14日)',
    macd DECIMAL(10,4) COMMENT 'MACD',
    macd_signal DECIMAL(10,4) COMMENT 'MACDシグナル',
    macd_histogram DECIMAL(10,4) COMMENT 'MACDヒストグラム',
    sma_5 DECIMAL(10,2) COMMENT '5日移動平均',
    sma_25 DECIMAL(10,2) COMMENT '25日移動平均',
    sma_75 DECIMAL(10,2) COMMENT '75日移動平均',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    UNIQUE KEY unique_code_date (code, `date`),
    INDEX idx_code (code),
    INDEX idx_date (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='テクニカル指標';

-- ポートフォリオテーブル
CREATE TABLE portfolios (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    name VARCHAR(100) NOT NULL COMMENT '銘柄名',
    shares INT NOT NULL COMMENT '保有株数',
    purchase_price DECIMAL(10,2) NOT NULL COMMENT '購入価格',
    purchase_date DATE NOT NULL COMMENT '購入日',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ポートフォリオ';

-- ウォッチリストテーブル
CREATE TABLE watch_lists (
    id VARCHAR(26) PRIMARY KEY,
    code VARCHAR(10) NOT NULL COMMENT '銘柄コード',
    name VARCHAR(100) NOT NULL COMMENT '銘柄名',
    target_buy_price DECIMAL(10,2) COMMENT '目標買い価格',
    target_sell_price DECIMAL(10,2) COMMENT '目標売り価格',
    is_active BOOLEAN DEFAULT TRUE COMMENT 'アクティブフラグ',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    UNIQUE KEY unique_code (code),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ウォッチリスト';