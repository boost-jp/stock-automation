# システムアーキテクチャ

## 概要

株式自動化システムは、マイクロサービスアーキテクチャを採用したGo製のバックエンドシステムです。Yahoo Finance APIからの株価データ収集、テクニカル分析、ポートフォリオ分析、およびSlack通知機能を提供します。

## システム構成図

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Data Collector │    │    Scheduler    │    │ Daily Reporter  │
│                 │    │                 │    │                 │
│  Yahoo Finance  │◄───┤  Cron Jobs      │───►│  Daily Reports  │
│  API Integration│    │  Task Queue     │    │  Notifications  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Database Layer                            │
│                        MySQL                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   Stocks    │  │   Prices    │  │   Reports   │            │
│  │   Table     │  │   Table     │  │   Table     │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Analysis     │    │   Notification  │    │   Web API       │
│                 │    │                 │    │                 │
│  Technical      │    │  Slack          │    │  REST Endpoints │
│  Portfolio      │    │  Integration    │    │  JSON Response  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## コンポーネント詳細

### 1. Data Collector (データ収集層)
- **役割**: Yahoo Finance APIからの株価データ取得
- **実装**: `internal/api/data_collector.go`
- **機能**:
  - リアルタイム株価取得
  - 履歴データ取得
  - レート制限対応
  - エラーハンドリング

### 2. Scheduler (スケジューラー)
- **役割**: 定期実行タスクの管理
- **実装**: `internal/api/scheduler.go`
- **機能**:
  - Cronベースの定期実行
  - タスクキューイング
  - 重複実行防止

### 3. Analysis Engine (分析エンジン)
- **役割**: テクニカル分析とポートフォリオ分析
- **実装**: 
  - `internal/analysis/technical.go`
  - `internal/analysis/portfolio.go`
- **機能**:
  - 移動平均計算
  - RSI指標計算
  - ポートフォリオパフォーマンス分析
  - リスク評価

### 4. Daily Reporter (日次レポート)
- **役割**: 日次レポート生成と通知
- **実装**: `internal/api/daily_reporter.go`
- **機能**:
  - 日次サマリー生成
  - パフォーマンスレポート
  - アラート生成

### 5. Notification System (通知システム)
- **役割**: 外部システムへの通知
- **実装**: `internal/notification/slack.go`
- **機能**:
  - Slack通知
  - 日本語メッセージ対応
  - 通知フォーマット管理

### 6. Database Layer (データベース層)
- **役割**: データ永続化
- **実装**: `internal/database/`
- **技術**: MySQL
- **機能**:
  - 株価データ保存
  - 分析結果保存
  - トランザクション管理

## データフロー

### 1. データ収集フロー
```
Yahoo Finance API → Data Collector → Database → Analysis Engine
```

### 2. 分析フロー
```
Database → Analysis Engine → Analysis Results → Database
```

### 3. 通知フロー
```
Scheduler → Daily Reporter → Analysis Results → Notification System → Slack
```

## デプロイメント構成

### 開発環境
- Docker Compose
- ローカルMySQL
- ホットリロード対応

### 本番環境（想定）
- Cloud Run
- Cloud SQL (MySQL)
- CI/CD パイプライン

## セキュリティ考慮事項

### 1. API Key管理
- 環境変数による秘匿情報管理
- `.env`ファイルの.gitignore登録

### 2. データベース接続
- 接続文字列の暗号化
- 最小権限の原則

### 3. 外部API通信
- HTTPS通信の強制
- タイムアウト設定

## パフォーマンス考慮事項

### 1. データベース最適化
- 適切なインデックス設定
- コネクションプーリング
- クエリ最適化

### 2. API制限対応
- レート制限の実装
- リトライ機構
- キャッシュ戦略

### 3. 並行処理
- Goroutineの活用
- チャネルによる通信
- コンテキストによるキャンセレーション

## 拡張性

### 1. 新しいデータソース追加
- プラグインアーキテクチャの採用
- インターフェースベースの設計

### 2. 新しい分析手法追加
- 分析エンジンのモジュール化
- 設定ベースの分析選択

### 3. 新しい通知チャネル追加
- 通知システムの抽象化
- マルチチャネル対応

## 監視・ログ

### 1. ログ戦略
- 構造化ログ（JSON形式）
- ログレベル管理
- 日本語対応

### 2. メトリクス収集
- パフォーマンスメトリクス
- ビジネスメトリクス
- エラーレート監視

### 3. ヘルスチェック
- アプリケーションヘルスチェック
- データベース接続チェック
- 外部API接続チェック