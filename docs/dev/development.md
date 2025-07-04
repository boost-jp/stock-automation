# 開発者ガイド

## 概要

株式自動化システムの開発環境構築と開発プロセスについて説明します。

## 開発環境構築

### 前提条件

- Go 1.24.4+
- Docker & Docker Compose
- MySQL 8.4+
- Git

### セットアップ手順

1. **リポジトリのクローン**
```bash
git clone <repository-url>
cd stock-automation
```

2. **開発ツールのインストール**
```bash
cd backend
make install-tools
```

3. **データベースの起動**
```bash
make docker-up
```

4. **アプリケーションの起動**
```bash
make dev
```

## 開発プロセス

### コード品質管理

#### 1. コードフォーマット
```bash
make fmt      # コードフォーマット
make lint     # リンターチェック
make vet      # go vet実行
```

#### 2. セキュリティチェック
```bash
make security # gosecによるセキュリティスキャン
```

#### 3. 全体チェック
```bash
make check    # 全コード品質チェック実行
```

### テスト戦略

#### 1. 単体テスト
```bash
make test              # 単体テスト実行
make test-coverage     # カバレッジ付きテスト
```

#### 2. 統合テスト
```bash
make test-integration  # 統合テスト実行
make test-all         # 全テスト実行
```

### ビルドとデプロイ

#### 1. ローカルビルド
```bash
make build        # ローカルビルド
make build-linux  # Linux向けビルド
```

#### 2. 開発環境
```bash
make dev          # 開発環境起動
make hot-reload   # ホットリロード開発
```

## Git ワークフロー

### ブランチ戦略

- **main**: 本番リリース用
- **develop**: 開発統合用
- **feature/**: 機能開発用
- **hotfix/**: 緊急修正用

### コミット前チェック

Pre-commitフックが自動実行されます：

```bash
# Pre-commitインストール（初回のみ）
pip install pre-commit
pre-commit install

# 手動実行
pre-commit run --all-files
```

### コミット規約

Conventional Commitsに従ってください：

```
feat: 新機能の追加
fix: バグ修正
docs: ドキュメント変更
style: フォーマット変更
refactor: リファクタリング
test: テスト追加・修正
chore: その他の変更
```

## CI/CD パイプライン

### GitHub Actions

`.github/workflows/ci.yml`で以下が自動実行されます：

1. **コード品質チェック**
   - gofmt, go vet
   - golangci-lint
   - gosec セキュリティスキャン

2. **テスト**
   - 単体テスト（カバレッジ付き）
   - 統合テスト（MySQL使用）

3. **ビルド**
   - バイナリビルド
   - アーティファクト保存

### トリガー条件

- **Push**: `main`, `develop`ブランチ
- **Pull Request**: `main`, `develop`ブランチ向け

## プロジェクト構造

```
backend/
├── cmd/              # エントリーポイント
│   ├── main.go      # メインアプリケーション
│   └── test_*.go    # テスト用スクリプト
├── internal/         # プライベートパッケージ
│   ├── analysis/    # 分析ロジック
│   ├── api/         # API層
│   ├── database/    # データベース層
│   ├── models/      # データモデル
│   └── notification/ # 通知システム
├── configs/         # 設定ファイル
├── docker/          # Docker設定
├── Makefile         # ビルドスクリプト
├── .golangci.yml    # Linter設定
├── .gosec.json      # セキュリティスキャン設定
└── go.mod          # Go依存関係
```

## 設定管理

### 環境変数

以下の環境変数を設定してください：

```bash
# データベース設定
DB_HOST=localhost
DB_PORT=3309
DB_NAME=stock_automation
DB_USER=root
DB_PASSWORD=password

# 外部API設定
YAHOO_FINANCE_API_KEY=your_api_key
SLACK_WEBHOOK_URL=your_webhook_url

# アプリケーション設定
APP_ENV=development
LOG_LEVEL=debug
```

### 設定ファイル

`configs/config.yaml`で各種設定を管理：

```yaml
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DB_NAME}
  user: ${DB_USER}
  password: ${DB_PASSWORD}

api:
  yahoo_finance:
    api_key: ${YAHOO_FINANCE_API_KEY}
    rate_limit: 100

notification:
  slack:
    webhook_url: ${SLACK_WEBHOOK_URL}
```

## 開発のベストプラクティス

### 1. コーディング規約

- **Go標準**: `gofmt`に従う
- **命名規約**: Go conventionsに従う
- **エラーハンドリング**: 明示的なエラー処理
- **ロギング**: 構造化ログ（JSON形式）
- **日本語対応**: UTF-8エンコーディング

### 2. テスト指針

- **カバレッジ**: 80%以上を目標
- **テストファイル**: `*_test.go`
- **テストデータ**: `testdata/`ディレクトリ
- **モック**: 外部依存のモック化

### 3. セキュリティ

- **秘匿情報**: 環境変数で管理
- **入力検証**: 全外部入力の検証
- **SQLインジェクション**: ORM使用
- **XSS対策**: 適切なエスケープ

### 4. パフォーマンス

- **データベース**: 適切なインデックス
- **並行処理**: Goroutineの活用
- **メモリ管理**: リークの防止
- **API制限**: レート制限の実装

## トラブルシューティング

### よくある問題

#### 1. データベース接続エラー
```bash
# Docker起動確認
docker-compose ps

# ログ確認
docker-compose logs mysql
```

#### 2. Go依存関係エラー
```bash
# モジュール同期
go mod tidy
go mod download
```

#### 3. テスト失敗
```bash
# 詳細ログで実行
go test -v ./...

# 個別テスト実行
go test -v ./internal/models
```

#### 4. Linterエラー
```bash
# 自動修正可能な問題
make fmt

# 詳細レポート
golangci-lint run --verbose
```

## デバッグ

### 1. ログ設定

```go
// 開発時
logrus.SetLevel(logrus.DebugLevel)

// 本番時
logrus.SetLevel(logrus.InfoLevel)
```

### 2. デバッガー使用

```bash
# Delveインストール
go install github.com/go-delve/delve/cmd/dlv@latest

# デバッグ実行
dlv debug cmd/main.go
```

### 3. プロファイリング

```bash
# CPUプロファイル
go test -cpuprofile=cpu.prof -bench=.

# メモリプロファイル
go test -memprofile=mem.prof -bench=.
```

## 参考リンク

- [Go Documentation](https://golang.org/doc/)
- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://securecodewarrior.github.io/gosec/)
- [GORM Documentation](https://gorm.io/)
- [Testify](https://github.com/stretchr/testify)

## お困りの際は

開発に関する質問や問題は、以下の方法でサポートを受けてください：

1. **Issue作成**: GitHubのIssueとして報告
2. **ドキュメント確認**: このガイドと関連ドキュメント
3. **コード参照**: 既存の実装例を参考に