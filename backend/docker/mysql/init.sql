-- Stock Automation Database Initialization

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS stock_automation CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the database
USE stock_automation;

-- Create sample watch list data
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('7203', 'トヨタ自動車', 2000.00, 2500.00, true, NOW(), NOW()),
('6758', 'ソニーグループ', 8000.00, 12000.00, true, NOW(), NOW()),
('9984', 'ソフトバンクグループ', 5000.00, 7000.00, true, NOW(), NOW()),
('8306', '三菱UFJフィナンシャル・グループ', 700.00, 1000.00, true, NOW(), NOW()),
('6501', '日立製作所', 6000.00, 8000.00, true, NOW(), NOW());

-- Create sample portfolio data
INSERT INTO portfolios (code, name, shares, purchase_price, purchase_date, created_at, updated_at) VALUES
('7203', 'トヨタ自動車', 100, 2100.00, '2023-01-15', NOW(), NOW()),
('6758', 'ソニーグループ', 50, 9500.00, '2023-02-20', NOW(), NOW()),
('9984', 'ソフトバンクグループ', 200, 5500.00, '2023-03-10', NOW(), NOW());

-- Create indexes for better performance
CREATE INDEX idx_stock_prices_code_timestamp ON stock_prices (code, timestamp);
CREATE INDEX idx_technical_indicators_code_timestamp ON technical_indicators (code, timestamp);

-- Show table structure
DESCRIBE watch_lists;
DESCRIBE portfolios;
DESCRIBE stock_prices;
DESCRIBE technical_indicators;