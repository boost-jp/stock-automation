# システム動作フロー - シーケンス図

## 概要

株式自動化システムの主要な動作フローをシーケンス図で可視化し、各コンポーネントの連携と処理の流れを明確化します。

## 🔄 メインフロー

### 1. 📊 日次データ収集・分析フロー

```mermaid
sequenceDiagram
    participant Scheduler as スケジューラー
    participant DataCollector as データ収集
    participant YahooAPI as Yahoo Finance API
    participant Database as データベース
    participant Analysis as 分析エンジン
    participant Reporter as レポーター
    participant Slack as Slack通知

    Note over Scheduler: 平日 9:00-15:00
    Scheduler->>DataCollector: 株価データ収集開始
    
    loop 監視銘柄ごと
        DataCollector->>YahooAPI: 株価データリクエスト
        YahooAPI-->>DataCollector: 株価データレスポンス
        DataCollector->>Database: STOCK_PRICE テーブル保存
    end
    
    DataCollector->>Analysis: テクニカル分析開始
    Analysis->>Database: 株価履歴データ取得
    Database-->>Analysis: 履歴データ
    
    loop 各銘柄
        Analysis->>Analysis: 移動平均・RSI・MACD計算
        Analysis->>Database: TECHNICAL_INDICATOR保存
    end
    
    Note over Scheduler: 毎日 18:00
    Scheduler->>Reporter: 日次レポート生成
    Reporter->>Database: ポートフォリオデータ取得
    Reporter->>Database: 最新株価データ取得
    Database-->>Reporter: データセット
    
    Reporter->>Reporter: 損益計算・レポート整形
    Reporter->>Slack: 日次レポート送信
    Slack-->>Reporter: 送信完了
```

**このフローの目的**: 
- 定期的な株価データ収集の自動化
- テクニカル指標の自動計算
- 日次投資状況レポートの自動配信

### 2. 🚨 リアルタイム監視・アラートフロー

```mermaid
sequenceDiagram
    participant Scheduler as スケジューラー
    participant Monitor as 監視システム
    participant Database as データベース
    participant Analysis as 分析エンジン
    participant Slack as Slack通知

    Note over Scheduler: 市場時間中 1分ごと
    Scheduler->>Monitor: 監視チェック開始
    
    Monitor->>Database: WATCH_LIST取得
    Monitor->>Database: 最新STOCK_PRICE取得
    Database-->>Monitor: 監視銘柄・価格データ
    
    loop 監視銘柄ごと
        Monitor->>Monitor: 価格条件チェック
        alt 買い条件達成
            Monitor->>Analysis: テクニカル指標確認
            Analysis-->>Monitor: "Strong Buy"シグナル
            Monitor->>Slack: 🔥買いアラート送信
        else 売り条件達成
            Monitor->>Analysis: テクニカル指標確認  
            Analysis-->>Monitor: "Strong Sell"シグナル
            Monitor->>Slack: 💰売りアラート送信
        else 条件未達成
            Note over Monitor: 待機継続
        end
    end
```

**このフローの目的**:
- 投資機会の自動検出
- 売買タイミングの即座通知
- 人間の判断をサポート

### 3. 📈 ポートフォリオ分析フロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant API as Web API
    participant Portfolio as ポートフォリオ分析
    participant Database as データベース
    participant Formatter as レポート整形

    User->>API: ポートフォリオ状況確認
    API->>Portfolio: 分析リクエスト
    
    Portfolio->>Database: PORTFOLIO データ取得
    Portfolio->>Database: 最新 STOCK_PRICE取得
    Database-->>Portfolio: データセット
    
    loop 保有銘柄ごと
        Portfolio->>Portfolio: 現在評価額計算
        Portfolio->>Portfolio: 損益計算
        Portfolio->>Portfolio: 利回り計算
    end
    
    Portfolio->>Portfolio: ポートフォリオ全体集計
    Portfolio->>Formatter: レポート整形依頼
    
    Note over Formatter: sprintf関数でフォーマット
    Formatter->>Formatter: 日本語レポート生成
    
    Formatter-->>Portfolio: 整形済みレポート
    Portfolio-->>API: 分析結果
    API-->>User: レポート表示
```

**このフローの目的**:
- 投資パフォーマンスの可視化
- 損益状況の正確な把握
- 投資判断材料の提供

### 4. ⚙️ システム初期化・設定フロー

```mermaid
sequenceDiagram
    participant Admin as 管理者
    participant System as システム
    participant Database as データベース
    participant Config as 設定管理
    participant External as 外部API

    Admin->>System: システム起動
    System->>Config: 設定ファイル読み込み
    Config-->>System: 環境変数・設定値
    
    System->>Database: DB接続確立
    Database-->>System: 接続成功
    
    System->>Database: テーブル存在確認
    alt テーブル未作成
        System->>Database: マイグレーション実行
        Database-->>System: テーブル作成完了
    end
    
    System->>External: Yahoo Finance API接続テスト
    External-->>System: 接続成功
    
    System->>System: スケジューラー開始
    System-->>Admin: システム準備完了
```

### 5. 🚨 エラーハンドリング・回復フロー

```mermaid
sequenceDiagram
    participant System as システム
    participant DataCollector as データ収集
    participant YahooAPI as Yahoo Finance API
    participant Logger as ログシステム
    participant Slack as Slack通知
    participant Recovery as 回復処理

    System->>DataCollector: データ収集実行
    DataCollector->>YahooAPI: API リクエスト
    
    alt API エラー（レート制限）
        YahooAPI-->>DataCollector: 429 Too Many Requests
        DataCollector->>Logger: エラーログ記録
        DataCollector->>Recovery: 指数バックオフ待機
        Recovery->>DataCollector: リトライ実行
        DataCollector->>YahooAPI: 再リクエスト
        YahooAPI-->>DataCollector: 成功レスポンス
    else ネットワークエラー
        YahooAPI-->>DataCollector: Network Timeout
        DataCollector->>Logger: エラーログ記録
        DataCollector->>Recovery: 3回リトライ
        alt リトライ成功
            Recovery->>DataCollector: 回復完了
        else リトライ失敗
            Recovery->>Slack: 🚨システムエラー通知
            Recovery->>System: 緊急停止判定
        end
    end
```

## 📋 シーケンス図から見える実装のポイント

### 🎯 現在の実装状況と必要な作業

#### ✅ 実装済み
- データベース設計・接続
- Yahoo Finance API 連携
- テクニカル指標計算ロジック
- 基本的なエラーハンドリング

#### 🚧 要実装（sprintf関数等）

**1. Reporter の sprintf関数**
```go
// 現在: プレースホルダー状態
func sprintf(format string, args ...interface{}) string {
    // TODO: 実装が必要
}

// 必要な実装: 日本語レポート整形
func sprintf(format string, args ...interface{}) string {
    switch format {
    case "現在価値: ¥%,.0f\n":
        return fmt.Sprintf("現在価値: ¥%,.0f\n", args[0])
    // ... 他のフォーマット対応
    }
}
```

**2. DataCollector の完全実装**
```go
// 必要な機能
- Yahoo Finance API からの安定した株価取得
- エラー時のリトライ機能
- レート制限対応
- データベース保存処理
```

**3. スケジューラーとの統合**
```go
// 必要な機能  
- 定時実行の設定
- 各コンポーネントとの連携
- エラー時の通知
```

### 🔄 データフローの重要性

これらのシーケンス図により、各実装が**なぜ必要か**が明確になります：

1. **sprintf関数**: ポートフォリオ分析結果を人間が読みやすい形式でSlack通知するため
2. **DataCollector**: 投資判断の基盤となる正確な株価データを自動収集するため  
3. **スケジューラー**: 人間の操作なしに24時間自動で投資監視を行うため

これで実装内容の目的と全体の中での位置づけが明確になりました！