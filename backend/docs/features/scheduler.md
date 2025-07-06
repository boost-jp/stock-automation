# スケジューラー機能

## 概要

株価データの自動収集、レポート生成、データクリーンアップを定期的に実行するスケジューラー機能です。

## 機能

### 1. 定期実行タスク

#### 株価データ更新（5分毎）
- 市場開場時間中のみ実行
- 全銘柄の最新株価を取得
- 実行ログをデータベースに記録

#### 設定更新（30分毎）
- ウォッチリストの更新
- ポートフォリオ設定の更新

#### デイリーレポート（平日18:00）
- ポートフォリオサマリーをSlackに送信
- 月曜日から金曜日のみ実行
- 包括的なレポートフォーマット

#### データクリーンアップ（毎日2:00）
- 365日以上前の古いデータを削除
- データベースの最適化

### 2. 市場時間判定

日本市場の取引時間を考慮：
- 前場: 9:00 - 11:30
- 後場: 12:30 - 15:00
- 土日祝日は除外

### 3. 実行ログ機能

全てのスケジュールタスクの実行履歴を記録：
- タスク名
- 実行開始時刻
- 実行完了時刻
- 実行時間（ミリ秒）
- エラーメッセージ（失敗時）

## 使用方法

### スケジューラーの起動

```bash
# スケジューラーを起動
./stock-automation scheduler

# または
./stock-automation run
```

### 実行ログの確認

```bash
# 最近の実行ログを表示
./stock-automation logs
```

### 手動実行

```bash
# データ収集を即座に実行
./stock-automation collect

# デイリーレポートを即座に送信
./stock-automation report
```

## データベーススキーマ

スケジューラーログは`scheduler_logs`テーブルに保存されます：

```sql
CREATE TABLE scheduler_logs (
    id SERIAL PRIMARY KEY,
    task_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error_message TEXT,
    duration_ms INT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 設定

スケジューラーはJST（日本標準時）で動作します。タイムゾーンは自動的に設定されます。

## エラーハンドリング

- 各タスクのエラーは個別にログに記録
- エラーが発生してもスケジューラーは停止しない
- リトライ機能は各タスク内で実装

## モニタリング

### ログ出力
- 標準出力にタスクの開始・完了をログ出力
- エラー発生時は詳細なエラーメッセージを出力

### データベース確認
```sql
-- 最近の実行ログを確認
SELECT * FROM scheduler_logs 
ORDER BY started_at DESC 
LIMIT 20;

-- 特定タスクの履歴を確認
SELECT * FROM scheduler_logs 
WHERE task_name = 'daily_report' 
ORDER BY started_at DESC;

-- エラーログのみ確認
SELECT * FROM scheduler_logs 
WHERE status = 'failed' 
ORDER BY started_at DESC;
```

## トラブルシューティング

### スケジューラーが動作しない
1. ログレベルをdebugに設定して詳細を確認
2. データベース接続を確認
3. 環境変数が正しく設定されているか確認

### レポートが送信されない
1. Slack Webhook URLが設定されているか確認
2. 平日18:00に正しく実行されているかログを確認
3. ネットワーク接続を確認