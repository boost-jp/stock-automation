# Stock Automation System

Go言語で実装した株価データ収集・分析・通知システム（クリーンアーキテクチャ版）

## 概要

Yahoo Finance APIを使用して株価データを収集し、MySQL 8データベースに保存。
テクニカル指標を計算して投資判断を行い、Slack通知で結果を配信するシステムです。

クリーンアーキテクチャを採用し、ビジネスロジックを中心に据えた保守性の高い設計を実現しています。

## 主な機能

- 📊 **リアルタイム株価データ収集**
- 💾 **MySQL 8データベースでの高速データ保存**
- 📈 **テクニカル指標計算（MA、RSI、MACD）**
- 🔔 **Slack通知による価格アラート**
- 📋 **ポートフォリオ管理・損益計算**
- ⏰ **市場時間に合わせた自動実行**
- 🛠️ **CLIによる対話的操作**

## システム要件

- Go 1.24以上
- Docker & Docker Compose
- MySQL 8.0
- インターネット接続

## アーキテクチャ

### クリーンアーキテクチャの層構造

```
┌─────────────────────────────────────────────────┐
│              Interface Layer                     │
│  - CLI Commands                                 │
│  - Scheduler                                    │
│  - Dependency Injection Container               │
└─────────────────────┬───────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────┐
│              Use Case Layer                     │
│  - CollectDataUseCase                          │
│  - PortfolioReportUseCase                      │
│  - TechnicalAnalysisUseCase                    │
└─────────────────────┬───────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────┐
│              Domain Layer                       │
│  - Entities (Portfolio, StockPrice, etc.)      │
│  - Domain Services                             │
│  - Business Rules                              │
└─────────────────────┬───────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────┐
│           Infrastructure Layer                  │
│  - Repository Implementation                    │
│  - External API Clients                        │
│  - Database Connection                         │
│  - Notification Services                       │
└─────────────────────────────────────────────────┘
```

## クイックスタート

### 1. 環境構築

```bash
# リポジトリクローン
git clone <repository-url>
cd stock-automation/backend

# 依存パッケージのインストール
go mod tidy

# Docker Composeでデータベース起動
docker compose up -d

# 環境変数設定
export SLACK_WEBHOOK_URL="your_slack_webhook_url"
export TEST_DB_PORT=3309
```

### 2. 実行

```bash
# スケジューラーを起動（デフォルト）
go run cmd/main.go

# 即座にデータ収集を実行
go run cmd/main.go collect

# 日次レポートを送信
go run cmd/main.go report

# ポートフォリオを表示
go run cmd/main.go portfolio list

# ヘルプを表示
go run cmd/main.go help
```

## プロジェクト構造

```
stock-automation/backend/
├── cmd/
│   └── main.go                 # エントリーポイント
├── app/
│   ├── domain/                 # ドメイン層
│   │   ├── models/            # エンティティ
│   │   └── services/          # ドメインサービス
│   ├── usecase/               # ユースケース層
│   │   ├── collect_data.go
│   │   ├── portfolio_report.go
│   │   └── technical_analysis.go
│   ├── infrastructure/        # インフラストラクチャ層
│   │   ├── client/           # 外部API
│   │   ├── config/           # 設定管理
│   │   ├── database/         # DB接続
│   │   ├── notification/     # 通知
│   │   └── repository/       # リポジトリ実装
│   ├── interfaces/            # インターフェース層
│   │   ├── cli.go            # CLIコマンド
│   │   ├── container.go      # DI コンテナ
│   │   └── scheduler.go      # スケジューラー
│   └── testutil/              # テストユーティリティ
├── tests/
│   ├── unit/                  # ユニットテスト
│   └── integration/           # 統合テスト
├── configs/
│   └── config.yaml            # 設定ファイル
├── schema.sql                 # データベーススキーマ
├── go.mod
├── go.sum
├── Makefile                   # ビルド・テストコマンド
└── README.md
```

## 設定

### 環境変数

```bash
# Slack通知
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
export SLACK_CHANNEL="#general"
export SLACK_USERNAME="Stock Bot"

# データベース
export DB_HOST="localhost"
export DB_PORT="3309"
export DB_USER="root"
export DB_PASSWORD="password"
export DB_NAME="stock_automation"

# テスト用データベース
export TEST_DB_HOST="localhost"
export TEST_DB_PORT="3309"
export TEST_DB_USER="root"
export TEST_DB_PASSWORD="password"
```

### 監視銘柄の追加

CLIを使用（今後実装予定）:
```bash
go run cmd/main.go watchlist add 7203 "トヨタ自動車"
```

または直接データベースに追加:
```sql
INSERT INTO watch_lists (id, code, name, target_buy_price, target_sell_price, is_active) 
VALUES (ULID(), '7203', 'トヨタ自動車', 2000.00, 2500.00, true);
```

## 開発

### テストの実行

```bash
# 全テストを実行
make test

# ユニットテストのみ
go test ./app/domain/...

# 統合テスト込み
make test-integration

# カバレッジレポート生成
make test-coverage
```

### ビルド

```bash
# アプリケーションをビルド
make build

# Linuxビルド（デプロイ用）
make build-linux
```

### コード品質チェック

```bash
# フォーマット
make fmt

# リント
make lint

# 全てのチェック
make check
```

## 運用

### ログ監視

アプリケーションはlogrusを使用して構造化ログを出力します。

```bash
# JSON形式でログを確認
go run cmd/main.go --log-level=debug 2>&1 | jq '.'
```

### データベース管理

```bash
# マイグレーション実行
make migrate

# テストDB作成
make test-db-setup

# SQLBoilerモデル生成
make gen-sqlboiler
```

### Docker操作

```bash
# 開発環境起動
make docker-up

# 停止
make docker-down

# ログ確認
make docker-logs
```

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   ```bash
   # Docker コンテナの状態確認
   docker ps
   
   # 環境変数の確認
   echo $DB_PORT
   ```

2. **Yahoo Finance API エラー**
   - レート制限を確認（10リクエスト/秒）
   - ネットワーク接続を確認

3. **Slack通知が届かない**
   - Webhook URLの有効性を確認
   - 環境変数が正しく設定されているか確認

## アーキテクチャの詳細

### ドメイン層
- ビジネスロジックの中核
- 外部依存を持たない純粋なGoコード
- エンティティとドメインサービスで構成

### ユースケース層
- アプリケーションのユースケースを実装
- ドメイン層とインフラ層を協調
- トランザクション管理

### インフラストラクチャ層
- 外部システムとの接続を実装
- リポジトリパターンでデータアクセスを抽象化
- 依存性逆転の原則を適用

### インターフェース層
- ユーザーとシステムの接点
- CLIコマンドとスケジューラーを提供
- 依存性注入コンテナで全体を統合

## ライセンス

MIT License

## 貢献

プルリクエスト歓迎です。バグ報告や機能要望はIssueでお願いします。

### 開発ガイドライン
- クリーンアーキテクチャの原則を守る
- テストを必ず書く（カバレッジ80%以上）
- `make check`が通ることを確認
- コミットメッセージは[Conventional Commits](https://www.conventionalcommits.org/)に従う