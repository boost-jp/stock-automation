-- 高流動性銘柄の追加
-- 既存の監視銘柄に加えて、セクター分散された流動性の高い銘柄を追加

-- 自動車セクター
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('7267', 'ホンダ', 3200, 4000, true, NOW(), NOW());

-- テクノロジー・エレクトロニクス
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('6861', 'キーエンス', 45000, 60000, true, NOW(), NOW()),
('7974', '任天堂', 5500, 7500, true, NOW(), NOW());

-- 金融セクター
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('8316', '三井住友フィナンシャルグループ', 4500, 6000, true, NOW(), NOW()),
('8411', 'みずほフィナンシャルグループ', 1600, 2200, true, NOW(), NOW());

-- 通信・IT
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('9432', 'NTT', 120, 160, true, NOW(), NOW());

-- 重電・機械
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('6503', '三菱電機', 1400, 1800, true, NOW(), NOW());

-- 医薬品・ヘルスケア
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('4502', '武田薬品工業', 3500, 4500, true, NOW(), NOW()),
('4568', '第一三共', 4000, 5500, true, NOW(), NOW());

-- 小売・サービス
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active, created_at, updated_at) VALUES
('9983', 'ファーストリテイリング', 8000, 12000, true, NOW(), NOW());

-- 確認用クエリ
SELECT 
    code,
    name,
    target_buy_price,
    target_sell_price,
    CASE 
        WHEN code IN ('7203', '7267') THEN '自動車'
        WHEN code IN ('6758', '6861', '7974') THEN 'テクノロジー'
        WHEN code IN ('8306', '8316', '8411') THEN '金融'
        WHEN code IN ('9984', '9432') THEN '通信・IT'
        WHEN code IN ('6501', '6503') THEN '重電・機械'
        WHEN code IN ('4502', '4568') THEN '医薬品'
        WHEN code = '9983' THEN '小売'
    END as sector
FROM watch_lists 
WHERE is_active = true 
ORDER BY sector, code;