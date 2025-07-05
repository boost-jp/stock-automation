# Stock Automation System

Go言語で実装した株価データ収集・分析・通知システム

## 概要

Yahoo Finance APIを使用して株価データを収集し、MySQL 8データベースに保存。
テクニカル指標を計算して投資判断を行い、Slack通知で結果を配信するシステムです。

## 主な機能

- 📊 **リアルタイム株価データ収集**
- 💾 **MySQL 8データベースでの高速データ保存**
- 📈 **テクニカル指標計算（MA、RSI、MACD）**
- 🔔 **Slack通知による価格アラート**
- 📋 **ポートフォリオ管理・損益計算**
- ⏰ **市場時間に合わせた自動実行**

## システム要件

- Go 1.19以上
- Docker & Docker Compose
- MySQL 8.0
- インターネット接続

## クイックスタート

### 1. 環境構築

```bash
# リポジトリクローン
git clone <repository-url>
cd stock-automation

# 依存パッケージのインストール
go mod tidy

# Docker Composeでデータベース起動
cd docker
docker-compose up -d

# 環境変数設定
export SLACK_WEBHOOK_URL="your_slack_webhook_url"
```

### 2. 実行

```bash
# プログラム起動
go run cmd/main.go
```

## プロジェクト構造

```
stock-automation/
├── cmd/
│   └── main.go                 # メインプログラム
├── app/
│   ├── api/                    # 外部API連携
│   │   ├── yahoo_finance.go    # Yahoo Finance API
│   │   ├── data_collector.go   # データ収集
│   │   └── scheduler.go        # スケジューラー
│   ├── database/               # データベース操作
│   │   ├── database.go         # DB接続
│   │   └── stock_operations.go # 株価データ操作
│   ├── notification/           # 通知システム
│   │   └── slack.go            # Slack通知
│   └── models/                 # データ構造
│       └── stock.go            # 株価モデル
├── configs/
│   └── config.yaml            # 設定ファイル
├── docker/
│   ├── docker-compose.yml     # Docker Compose設定
│   └── mysql/
│       └── init.sql           # MySQL初期化
├── go.mod
├── go.sum
└── README.md
```

## 設定

### 環境変数

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
```

### 監視銘柄の追加

データベースに直接追加するか、プログラムで管理:

```sql
INSERT INTO watch_lists (code, name, target_buy_price, target_sell_price, is_active) 
VALUES ('7203', 'トヨタ自動車', 2000.00, 2500.00, true);
```

## 主要コンポーネント

### データ収集 (Yahoo Finance API)
- 5分間隔でのリアルタイム価格取得
- 履歴データの一括取得
- データ品質チェック・バリデーション

### データベース (MySQL 8)
- 高速なデータ保存・検索
- 自動マイグレーション
- 古いデータの自動削除

### 通知システム (Slack)
- 価格アラート通知
- 日次レポート配信
- システム状態通知

### スケジューラー
- 市場時間の自動判定
- 定期的なデータ更新
- エラーハンドリング・リトライ

## 開発・拡張

### 新しい銘柄の追加

```go
// プログラムで監視銘柄を追加
watchList := models.WatchList{
    Code:            "6758",
    Name:            "ソニーグループ",
    TargetBuyPrice:  8000.00,
    TargetSellPrice: 12000.00,
    IsActive:        true,
}
```

### カスタム指標の実装

`app/analysis/` パッケージに新しい分析ロジックを追加可能

### 通知方法の追加

`app/notification/` パッケージに新しい通知方法を実装可能

## 運用

### ログ監視

```bash
# アプリケーションログ
tail -f logs/app.log

# データベースログ
docker-compose logs mysql
```

### データベース管理

```bash
# バックアップ
docker exec stock-automation-mysql mysqldump -u root -p stock_automation > backup.sql

# 復元
docker exec -i stock-automation-mysql mysql -u root -p stock_automation < backup.sql
```

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   - Docker Composeが起動しているか確認
   - 接続情報が正しいか確認

2. **Yahoo Finance API エラー**
   - ネットワーク接続を確認
   - レート制限に注意

3. **Slack通知が届かない**
   - Webhook URLが正しいか確認
   - 環境変数が設定されているか確認

## ライセンス

MIT License

## 貢献

プルリクエスト歓迎です。バグ報告や機能要望はIssueでお願いします。